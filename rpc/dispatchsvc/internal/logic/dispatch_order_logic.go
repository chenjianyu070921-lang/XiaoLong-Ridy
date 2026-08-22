package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/redisx"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// defaultDispatchTimeout 未配置 dispatchTimeoutSeconds 时的派单超时阈值。
const defaultDispatchTimeout = 60 * time.Second

// dispatchLockTTL 派单互斥锁过期时间，配合看门狗续期避免进程崩溃导致锁永久占用。
const dispatchLockTTL = 15 * time.Second

// dispatchLockKey 派单互斥锁的 Redis key（按订单维度）。
func dispatchLockKey(orderID uint64) string {
	return fmt.Sprintf("r:lock:dispatch:%d", orderID)
}

// lockDispatchOrder 获取该订单的派单互斥锁（Redis 分布式锁，跨实例生效），返回释放函数。
//
// 修复说明（P1-M4-3）：原实现用进程内 sync.Map 锁，多实例部署时各实例各自加锁，
// 并发派单会重复插入 dispatch_record。改为 Redis 分布式锁（带 owner token + 看门狗续期），
// 且始终返回非 nil 的 release 函数，未配置 Redis 时退化为空释放函数（单实例兼容）。
func (l *DispatchOrderLogic) lockDispatchOrder(ctx context.Context, orderID uint64) func() {
	noop := func() {}
	if l.svcCtx.Redis == nil {
		return noop
	}
	lk, err := redisx.TryLock(ctx, l.svcCtx.Redis, dispatchLockKey(orderID), dispatchLockTTL)
	if err != nil {
		// 获取锁失败时退化为不阻塞（由 DB 唯一约束/状态校验兜底），仅告警。
		l.Logger.Errorf("acquire dispatch lock failed, orderId=%d: %v", orderID, err)
		return noop
	}
	return lk.Release
}

type DispatchOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDispatchOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DispatchOrderLogic {
	return &DispatchOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 对指定订单执行派单（P0 为 mock 直派），返回写入的派单记录。
// 幂等保证：
//   - 已存在接单（Accepted）记录 → 直接返回已有记录，不重复派单；
//   - 已存在未超时的 Pending 记录（派单进行中）→ 直接返回已有记录；
//   - 记录全部为 Rejected/Timeout/Cancelled 或 Pending 已超时 → 历史记录标记超时后重新派单。
func (l *DispatchOrderLogic) DispatchOrder(in *proto.DispatchOrderRequest) (*proto.DispatchOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	// 按订单互斥串行化"检查-插入"临界区，防止并发派单插入重复记录。
	// 锁等待限制 5s：Redis 不可达或锁长期占有时避免派单请求无限挂起（P2-M4-7）。
	// 注意：cancel 必须 defer（而非立即调用），否则锁上下文被取消会导致分布式锁失效。
	lockCtx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()
	unlock := l.lockDispatchOrder(lockCtx, uint64(in.OrderId))
	defer unlock()

	timeout := time.Duration(l.svcCtx.Config.DispatchTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultDispatchTimeout
	}
	now := time.Now()

	// 状态复核：订单必须仍为待接单才可派单，防止取消/超时订单被竞态派单（P0-M4-1）。
	// 复核失败（下游不可用）时 fail-safe 拒绝派单；订单状态非待接单时直接返回空派单结果。
	if l.svcCtx.OrderStatusVerifier != nil {
		status, err := l.svcCtx.OrderStatusVerifier(l.ctx, in.OrderId)
		if err != nil {
			l.Logger.Errorf("verify order %d status failed, skip dispatch: %v", in.OrderId, err)
			return nil, err
		}
		if status != int32(constants.OrderStatusWaitAccept) {
			l.Logger.Infof("order %d status=%d is not wait_accept, skip dispatch", in.OrderId, status)
			return &proto.DispatchOrderResponse{OrderId: in.OrderId, List: []*proto.DispatchRecord{}}, nil
		}
	}

	if existing, total, err := l.svcCtx.DispatchRepository.ListByOrder(l.ctx, uint64(in.OrderId), 1, 100); err != nil {
		return nil, err
	} else if total > 0 {
		// 存在活跃派单（未超时的 Pending 或已接单）则幂等返回。
		if l.hasActiveDispatch(existing, now, timeout) {
			list := make([]*proto.DispatchRecord, 0, len(existing))
			for i := range existing {
				list = append(list, toProtoDispatchRecord(&existing[i]))
			}
			return &proto.DispatchOrderResponse{OrderId: in.OrderId, List: list}, nil
		}
		// 无活跃派单：先把历史超时 Pending 记录标记为超时，再重新派单。
		if _, err := l.svcCtx.DispatchRepository.MarkTimeoutByOrder(l.ctx, uint64(in.OrderId), now.Add(-timeout)); err != nil {
			return nil, err
		}
	}

	candidates, err := l.svcCtx.DispatchEngine.FindCandidates(
		l.ctx,
		uint64(in.OrderId),
		in.FromLongitude,
		in.FromLatitude,
		in.CarType,
		in.CityCode,
	)
	if err != nil {
		return nil, err
	}

	list := make([]*proto.DispatchRecord, 0, len(candidates))
	driverIDs := make([]int64, 0, len(candidates))
	for _, candidate := range candidates {
		record := &model.DispatchRecord{
			OrderId:      uint64(in.OrderId),
			DriverId:     candidate.DriverID,
			DispatchType: constants.DispatchTypeAuto,
			Status:       constants.DispatchStatusPending,
			MatchScore:   candidate.MatchScore,
			Remark:       "自动派单",
		}
		if err := l.svcCtx.DispatchRepository.Create(l.ctx, record); err != nil {
			return nil, err
		}
		list = append(list, toProtoDispatchRecord(record))
		driverIDs = append(driverIDs, int64(candidate.DriverID))
	}

	// 派单落库后发布通知事件，供司机端实时提醒。
	// 修复说明（P1-M4-4）：原实现发布失败仅记普通日志，Redis 瞬时抖动时事件静默丢失、
	// 司机端永远收不到派单导致订单卡死。此处增加一次重试，失败则输出告警级日志
	// （含可重放的关键字段），便于排查与人工/定时任务补偿，且不返回 error 以免上游误触发重复派单。
	if l.svcCtx.EventBus != nil && len(driverIDs) > 0 {
		payload, _ := json.Marshal(dispatchNewEvent{
			OrderId:       in.OrderId,
			DriverIds:     driverIDs,
			FromLongitude: in.FromLongitude,
			FromLatitude:  in.FromLatitude,
			CarType:       in.CarType,
			CityCode:      in.CityCode,
			DispatchedAt:  time.Now().Unix(),
		})
		if err := l.publishWithRetry(l.ctx, constants.TopicDispatchNew, payload); err != nil {
			l.Logger.Errorf("publish dispatch.new failed after retry, orderId=%d driverIds=%v: %v",
				in.OrderId, driverIDs, err)
		}
	}

	return &proto.DispatchOrderResponse{
		OrderId: in.OrderId,
		List:    list,
	}, nil
}

// hasActiveDispatch 判断记录中是否存在未超时的 Pending 或已接单的活跃派单。
func (l *DispatchOrderLogic) hasActiveDispatch(records []model.DispatchRecord, now time.Time, timeout time.Duration) bool {
	for i := range records {
		switch records[i].Status {
		case constants.DispatchStatusAccepted:
			return true
		case constants.DispatchStatusPending:
			if now.Sub(records[i].CreatedAt) < timeout {
				return true
			}
		}
	}
	return false
}

// publishWithRetry 发布事件，网络/中间件瞬时抖动时重试一次，降低事件静默丢失概率。
func (l *DispatchOrderLogic) publishWithRetry(ctx context.Context, topic string, payload []byte) error {
	if err := l.svcCtx.EventBus.Publish(ctx, topic, payload); err != nil {
		l.Logger.Errorf("publish %s first attempt failed: %v, retrying", topic, err)
		// 短暂退避后重试一次，避免上游误触发重复派单时不放大失败。
		time.Sleep(50 * time.Millisecond)
		return l.svcCtx.EventBus.Publish(ctx, topic, payload)
	}
	return nil
}

// toProtoDispatchRecord 将仓储派单记录转换为 proto 结构。
func toProtoDispatchRecord(record *model.DispatchRecord) *proto.DispatchRecord {
	return &proto.DispatchRecord{
		Id:           int64(record.Id),
		OrderId:      int64(record.OrderId),
		DriverId:     int64(record.DriverId),
		DispatchType: int32(record.DispatchType),
		Status:       int32(record.Status),
		MatchScore:   record.MatchScore,
		Remark:       record.Remark,
		CreatedAt:    record.CreatedAt.Unix(),
		UpdatedAt:    record.UpdatedAt.Unix(),
	}
}

// dispatchNewEvent 派单通知事件 payload（统一 snake_case，与消费端 DispatchNewEvent 对齐）。
type dispatchNewEvent struct {
	OrderId       int64   `json:"order_id"`
	DriverIds     []int64 `json:"driver_ids"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
	DispatchedAt  int64   `json:"dispatched_at"`
}
