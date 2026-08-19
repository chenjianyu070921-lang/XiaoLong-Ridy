package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
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
		order.Status,
	}, operatorType, reason, statusLog)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotCancelable
	}

	// 同步失效该订单的待派单记录，失败不阻断订单取消主流程。
	syncCancelDispatch(l.ctx, l.svcCtx.DispatchClient, order.Id, reason)

	return &proto.CancelOrderResponse{
		OrderId: in.OrderId,
		Status:  proto.OrderStatus_ORDER_STATUS_CANCELLED,
	}, nil
}

// syncCancelDispatch 订单取消后同步将派单记录置为已取消，失败仅记日志。
func syncCancelDispatch(ctx context.Context, dispatchClient dispatch.Dispatch, orderID uint64, reason string) {
	if dispatchClient == nil {
		return
	}
	_, err := dispatchClient.CancelDispatch(ctx, &dispatch.CancelDispatchRequest{
		OrderId: int64(orderID),
		Reason:  reason,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("sync cancel dispatch failed, order_id=%d err=%v", orderID, err)
	}
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
func canCancelByOperator(order *model.RideOrder, operatorType string, operatorID int64) bool {
	switch operatorType {
	case constants.OperatorUser:
		// 乘客只能取消未接单的订单；司机已接单后若需取消由司机/系统入口处理，避免“已接单又被乘客取消”的竞态。
		return order.Status == constants.OrderStatusWaitAccept && order.UserId == uint64(operatorID)
	case constants.OperatorDriver:
		return order.Status == constants.OrderStatusAccepted && order.DriverId == uint64(operatorID)
	case constants.OperatorSystem, constants.OperatorAdmin:
		return true
	default:
		return false
	}
}
