package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// 订单业务错误定义。
var (
	ErrInvalidOrderParams       = errors.New("invalid orderclient params")
	ErrCancelReasonRequired     = errors.New("cancel reason required")
	ErrOrderStatusNotCancelable = errors.New("orderclient status not cancelable")
	ErrCancelNotAllowed         = errors.New("operator not allowed to cancel this orderclient")
	ErrOrderStatusNotAllowed    = errors.New("orderclient status not allowed")
	ErrDriverNotMatched         = errors.New("driver not matched")
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCancelOrderLogic 创建取消订单逻辑对象。
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelOrder 校验取消方与订单状态，条件更新订单并写入取消日志。
func (l *CancelOrderLogic) CancelOrder(in *proto.CancelOrderRequest) (*proto.CancelOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	operatorType := strings.TrimSpace(in.OperatorType)
	if !validOperatorType(operatorType) {
		return nil, ErrInvalidOrderParams
	}
	if operatorType != constants.OperatorSystem && in.OperatorId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return nil, ErrCancelReasonRequired
	}

	// 订单级分布式锁：避免取消与接单/超时取消并发竞态。
	release, err := acquireOrderLock(l.ctx, l.svcCtx.Redis, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	defer release()

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}
	if !canCancelStatus(order.Status) {
		return nil, ErrOrderStatusNotCancelable
	}
	if !canCancelByOperator(order, operatorType, in.OperatorId) {
		return nil, ErrCancelNotAllowed
	}

	statusLog := &model.OrderStatusLog{
		FromStatus:   order.Status,
		ToStatus:     constants.OrderStatusCancelled,
		OperatorType: operatorType,
		OperatorId:   uint64(in.OperatorId),
		Remark:       reason,
	}
	ok, err := l.svcCtx.OrderRepository.Cancel(l.ctx, order.Id, []int8{
		constants.OrderStatusWaitAccept,
		constants.OrderStatusAccepted,
	}, operatorType, reason, statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotCancelable
	}

	// 已接单订单取消时，司机恢复可接单状态（P1-M4-8）。
	if order.DriverId > 0 {
		unmarkDriverBusy(l.ctx, l.svcCtx, order.DriverId)
	}

	// 取消成功后同步失效该订单的待派单记录，避免残留 Pending 被重派任务重复处理。
	syncCancelDispatch(l.ctx, l.svcCtx.DispatchClient, order.Id, reason)

	return &proto.CancelOrderResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_CANCELLED,
	}, nil
}

// validOperatorType 判断取消方类型是否合法。
func validOperatorType(operatorType string) bool {
	switch operatorType {
	case constants.OperatorUser, constants.OperatorDriver, constants.OperatorSystem, constants.OperatorAdmin:
		return true
	default:
		return false
	}
}

// canCancelStatus 判断订单状态是否允许取消。
func canCancelStatus(status int8) bool {
	return CanTransit(status, constants.OrderStatusCancelled)
}

// canCancelByOperator 校验取消方是否有权取消该订单。
// 未进入行程前允许乘客取消；进入行程后需由客服或系统处理。
func canCancelByOperator(order *model.RideOrder, operatorType string, operatorID int64) bool {
	switch operatorType {
	case constants.OperatorUser:
		return order.UserId == uint64(operatorID) && (order.Status == constants.OrderStatusWaitAccept || order.Status == constants.OrderStatusAccepted)
	case constants.OperatorDriver:
		return order.Status == constants.OrderStatusAccepted && order.DriverId == uint64(operatorID)
	case constants.OperatorSystem, constants.OperatorAdmin:
		return true
	default:
		return false
	}
}
