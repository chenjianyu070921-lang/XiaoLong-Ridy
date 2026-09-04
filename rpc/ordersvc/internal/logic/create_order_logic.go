package logic

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/keyutil"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	price "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// maxCreateOrderNoRetry 是订单号唯一索引冲突时的最大重试次数。
// 每次重试均重新生成订单号，限制重试次数避免异常情况下无限循环。
const maxCreateOrderNoRetry = 3

// dispatchRetryTask 描述一次待重试的派单请求，与 DispatchOrderRequest 所需参数一一对应。
// 序列化后作为 DispatchRetryQueueKey（Redis ZSet）的 member，由 job 服务扫描消费（P1-M4-2）。
type dispatchRetryTask struct {
	OrderId       int64   `json:"order_id"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
	Attempt       int     `json:"attempt"`
}

// enqueueDispatchRetry 将派单失败的订单写入延迟重试队列，退避间隔随尝试次数递增（5s/15s/45s）。
// Redis 不可用或写入失败时仅记录错误，不阻塞下单主流程。
func (l *CreateOrderLogic) enqueueDispatchRetry(task *dispatchRetryTask) {
	if l.svcCtx.Redis == nil {
		l.Logger.Errorf("dispatch retry queue disabled: redis not configured, order_id=%d", task.OrderId)
		return
	}
	if task.Attempt > constants.MaxDispatchRetryAttempt {
		l.Logger.Errorf("dispatch order %d retry exhausted (attempt=%d), keep in queue for manual intervention", task.OrderId, task.Attempt)
		return
	}
	payload, _ := json.Marshal(task)
	delay := time.Duration(5*int(math.Pow(3, float64(task.Attempt-1)))) * time.Second
	score := float64(time.Now().Add(delay).Unix())
	if err := l.svcCtx.Redis.ZAdd(l.ctx, constants.DispatchRetryQueueKey, redis.Z{Score: score, Member: string(payload)}).Err(); err != nil {
		l.Logger.Errorf("enqueue dispatch retry failed, order_id=%d: %v", task.OrderId, err)
		return
	}
	l.Logger.Errorf("dispatch order %d failed (attempt=%d), enqueued retry, next in %v", task.OrderId, task.Attempt, delay)
}

// NewCreateOrderLogic 创建订单逻辑对象。
func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type orderCreatedEvent struct {
	OrderId          int64   `json:"order_id"`
	OrderNo          string  `json:"order_no"`
	FromLongitude    float64 `json:"from_longitude"`
	FromLatitude     float64 `json:"from_latitude"`
	CarType          int32   `json:"car_type"`
	CityCode         string  `json:"city_code"`
	ExcludeDriverIds []int64 `json:"exclude_driver_ids"` // 改派/重派时排除的司机（如原司机）
}

// CreateOrder 校验参数并创建待接单订单，同时写入创建状态日志并发布订单创建事件。
func (l *CreateOrderLogic) CreateOrder(in *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	if err := validateCreateOrder(in); err != nil {
		return nil, err
	}
	if err := l.recordBlacklistOrderHit(uint64(in.UserId)); err != nil {
		return nil, err
	}

	// 服务端计价快照：优先调 pricesvc 复核，失败降级为入参预估价格。
	estimatedPriceCents := l.estimatePriceSnapshot(in)

	for attempt := 0; attempt < maxCreateOrderNoRetry; attempt++ {
		order := buildRideOrder(in, estimatedPriceCents)
		statusLog := &model.OrderStatusLog{
			FromStatus:   0,
			ToStatus:     constants.OrderStatusWaitAccept,
			OperatorType: constants.OperatorUser,
			OperatorId:   uint64(in.UserId),
			Remark:       "创建订单",
		}

		if err := l.svcCtx.OrderRepository.Create(l.ctx, order, statusLog); err != nil {
			if errors.Is(err, repository.ErrOrderNoExists) && attempt < maxCreateOrderNoRetry-1 {
				continue
			}
			return nil, err
		}

		// 下单即锁定优惠券（P1-订单-2）：保证"锁券->下单->支付->核销->失败回滚"的原子边界。
		// 锁券失败直接回滚订单，避免券被并发重复占用。
		if in.CouponId > 0 && l.svcCtx.CouponConsumer != nil {
			if lockErr := l.svcCtx.CouponConsumer.LockByOrder(l.ctx, uint64(in.UserId), uint64(in.CouponId), order.Id); lockErr != nil {
				l.Logger.Errorf("lock coupon %d for order %d failed: %v", in.CouponId, order.Id, lockErr)
				// 锁券失败：回滚已创建订单，返回明确错误。
				_, _ = l.svcCtx.OrderRepository.Cancel(l.ctx, order.Id,
					[]int8{constants.OrderStatusWaitAccept, constants.OrderStatusAccepted},
					constants.OperatorSystem, "优惠券锁定失败自动取消", &model.OrderStatusLog{
						FromStatus:   order.Status,
						ToStatus:     constants.OrderStatusCancelled,
						OperatorType: constants.OperatorSystem,
						OperatorId:   0,
						Remark:       "优惠券锁定失败自动取消",
					})
				return nil, repository.ErrCouponLockFailed
			}
		}

		// 优先发布 order.created 事件，由 order-event-consumer 触发派单。
		published := false
		if l.svcCtx.EventBus != nil {
			payload, _ := json.Marshal(orderCreatedEvent{
				OrderId:       int64(order.Id),
				OrderNo:       order.OrderNo,
				FromLongitude: in.FromLongitude,
				FromLatitude:  in.FromLatitude,
				CarType:       in.CarType,
				CityCode:      strings.TrimSpace(in.CityCode),
			})
			// 记录发送前的业务上下文，便于从前端下单请求追踪 Kafka 事件。
			l.Logger.Infof("order.created kafka publish start: orderId=%d orderNo=%s cityCode=%s carType=%d fromLng=%.6f fromLat=%.6f payloadBytes=%d",
				order.Id, order.OrderNo, strings.TrimSpace(in.CityCode), in.CarType, in.FromLongitude, in.FromLatitude, len(payload))
			if err := l.svcCtx.EventBus.Publish(l.ctx, constants.TopicOrderCreated, payload); err != nil {
				l.Logger.Errorf("order.created kafka publish failed: orderId=%d orderNo=%s topic=%s cityCode=%s error=%v",
					order.Id, order.OrderNo, constants.TopicOrderCreated, strings.TrimSpace(in.CityCode), err)
			} else {
				published = true
				l.Logger.Infof("order.created kafka publish succeeded: orderId=%d orderNo=%s topic=%s cityCode=%s",
					order.Id, order.OrderNo, constants.TopicOrderCreated, strings.TrimSpace(in.CityCode))
			}
		} else {
			l.Logger.Errorf("order.created kafka publish skipped: orderId=%d orderNo=%s reason=event_bus_not_configured",
				order.Id, order.OrderNo)
		}

		// 事件不可用时回退为同步直派，保证 demo 可跑通。
		if !published && l.svcCtx.DispatchClient != nil {
			if _, err := l.svcCtx.DispatchClient.DispatchOrder(l.ctx, &dispatch.DispatchOrderRequest{
				OrderId:       int64(order.Id),
				FromLongitude: in.FromLongitude,
				FromLatitude:  in.FromLatitude,
				CarType:       in.CarType,
				CityCode:      in.CityCode,
			}); err != nil {
				l.Logger.Errorf("dispatch order %d failed: %v", order.Id, err)
				// 派单失败补偿：写入延迟重试队列，由 job 扫描重试，避免订单永久卡在待接单（P1-M4-2）。
				l.enqueueDispatchRetry(&dispatchRetryTask{
					OrderId:       int64(order.Id),
					FromLongitude: in.FromLongitude,
					FromLatitude:  in.FromLatitude,
					CarType:       in.CarType,
					CityCode:      in.CityCode,
					Attempt:       1,
				})
			}
		}

		return &proto.CreateOrderResponse{
			OrderId:             int64(order.Id),
			OrderNo:             order.OrderNo,
			EstimatedPriceCents: estimatedPriceCents,
			Status:              proto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
			CreatedAt:           order.CreatedAt.Unix(),
		}, nil
	}

	return nil, ErrInvalidOrderParams
}

// recordBlacklistOrderHit 检查下单用户是否命中生效黑名单，并持久化下单场景审计记录。
// 黑名单查询失败仅记录日志并放行下单；已确认命中后写入失败返回错误，避免风险记录静默缺失。
func (l *CreateOrderLogic) recordBlacklistOrderHit(userID uint64) error {
	if l.svcCtx.RiskBlacklistRepository == nil {
		return nil
	}
	entry, err := l.svcCtx.RiskBlacklistRepository.FindActiveByTarget(l.ctx, "user", userID)
	if err != nil {
		l.Logger.Errorf("query order blacklist failed, user_id=%d: %v", userID, err)
		return nil
	}
	if entry == nil {
		return nil
	}
	return l.svcCtx.RiskBlacklistRepository.CreateHitRecord(l.ctx, &repository.BlacklistHitRecord{
		BlacklistID: entry.ID,
		TargetType:  "user",
		TargetID:    userID,
		Scene:       "order",
		RiskLevel:   3,
		HitReason:   entry.Reason,
		RequestID:   riskRequestID(l.ctx),
	})
}

// riskRequestID 从 gRPC 入站元数据提取请求链路 ID，并限制在运营表字段允许的长度内。
// 现有协议未强制携带该字段，缺失时返回空字符串以兼容历史调用。
func riskRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"x-request-id", "request-id", "trace-id"} {
		for _, value := range md.Get(key) {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if len(value) > 64 {
				return value[:64]
			}
			return value
		}
	}
	return ""
}

// buildRideOrder 根据 RPC 入参构造待接单订单，每次调用都会生成新的订单号。
func buildRideOrder(in *proto.CreateOrderRequest, estimatedPriceCents int64) *model.RideOrder {
	return &model.RideOrder{
		OrderNo:            keyutil.GenOrderID(),
		UserId:             uint64(in.UserId),
		DriverId:           0,
		CarType:            int8(in.CarType),
		CityCode:           strings.TrimSpace(in.CityCode),
		FromAddress:        strings.TrimSpace(in.FromAddress),
		FromLongitude:      in.FromLongitude,
		FromLatitude:       in.FromLatitude,
		ToAddress:          strings.TrimSpace(in.ToAddress),
		ToLongitude:        in.ToLongitude,
		ToLatitude:         in.ToLatitude,
		EstimatedDistanceM: int(in.EstimatedDistanceM),
		EstimatedDurationS: int(in.EstimatedDurationS),
		EstimatedPrice:     float64(estimatedPriceCents) / 100,
		Status:             constants.OrderStatusWaitAccept,
		// 价格与优惠券快照落库（P1-订单-1/2）：实付 = 预估 - 优惠，服务端以传参金额为准。
		CouponId:      in.CouponId,
		DiscountCents: in.DiscountCents,
		PayableCents:  estimatedPriceCents - in.DiscountCents,
	}
}

// defaultCityCode 兜底城市编码，与 api 网关默认值保持一致。
const defaultCityCode = "110000"

// estimatePriceSnapshot 确定订单要保存的金额快照。
// 乘客端已经根据用户优惠券计算并传入最终应付金额，订单服务必须优先保存该金额，
// 不能再次用未携带优惠券信息的 pricesvc 结果覆盖，否则会导致下单和支付恢复为原价。
// 只有调用方没有传入金额时，才使用 pricesvc 计算原始预估价作为兜底。
func (l *CreateOrderLogic) estimatePriceSnapshot(in *proto.CreateOrderRequest) int64 {
	if in.EstimatedPriceCents > 0 {
		return in.EstimatedPriceCents
	}
	if l.svcCtx == nil || l.svcCtx.PriceClient == nil {
		return in.EstimatedPriceCents
	}
	cityCode := strings.TrimSpace(in.CityCode)
	if cityCode == "" {
		cityCode = defaultCityCode
	}
	resp, err := l.svcCtx.PriceClient.EstimatePrice(l.ctx, &price.EstimatePriceRequest{
		UserId:    in.UserId,
		CityCode:  cityCode,
		CarType:   in.CarType,
		DistanceM: in.EstimatedDistanceM,
		DurationS: in.EstimatedDurationS,
	})
	if err != nil || resp == nil || resp.TotalCents <= 0 {
		l.Logger.Errorf("estimate price failed, fallback to input %d: %v", in.EstimatedPriceCents, err)
		return in.EstimatedPriceCents
	}
	return resp.TotalCents
}

// validateCreateOrder 校验创建订单入参。
func validateCreateOrder(in *proto.CreateOrderRequest) error {
	if in.UserId <= 0 ||
		in.CarType < 1 || in.CarType > 3 ||
		strings.TrimSpace(in.FromAddress) == "" ||
		strings.TrimSpace(in.ToAddress) == "" ||
		in.FromLongitude < -180 || in.FromLongitude > 180 ||
		in.FromLatitude < -90 || in.FromLatitude > 90 ||
		in.ToLongitude < -180 || in.ToLongitude > 180 ||
		in.ToLatitude < -90 || in.ToLatitude > 90 ||
		in.EstimatedDistanceM < 0 ||
		in.EstimatedDurationS < 0 ||
		in.EstimatedPriceCents < 0 {
		return ErrInvalidOrderParams
	}
	return nil
}
