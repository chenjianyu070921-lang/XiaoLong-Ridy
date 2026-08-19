package logic

import (
	"context"
	"encoding/json"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/model"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// defaultDispatchTimeout 未配置 dispatchTimeoutSeconds 时的派单超时阈值。
const defaultDispatchTimeout = 60 * time.Second

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

	timeout := time.Duration(l.svcCtx.Config.DispatchTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultDispatchTimeout
	}
	now := time.Now()

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

	// 派单落库后发布通知事件，供司机端实时提醒（失败不阻断派单主流程）。
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
		if err := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicDispatchNew, payload); err != nil {
			l.Logger.Errorf("publish dispatch.new failed: %v", err)
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

// dispatchNewEvent 派单通知事件 payload。
type dispatchNewEvent struct {
	OrderId       int64   `json:"orderId"`
	DriverIds     []int64 `json:"driverIds"`
	FromLongitude float64 `json:"fromLongitude"`
	FromLatitude  float64 `json:"fromLatitude"`
	CarType       int32   `json:"carType"`
	CityCode      string  `json:"cityCode"`
	DispatchedAt  int64   `json:"dispatchedAt"`
}
