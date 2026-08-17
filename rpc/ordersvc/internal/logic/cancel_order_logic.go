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
	ErrInvalidOrderParams       = errors.New("invalid order params")
	ErrCancelReasonRequired     = errors.New("cancel reason required")
	ErrOrderStatusNotCancelable = errors.New("order status not cancelable")
	ErrCancelNotAllowed         = errors.New("operator not allowed to cancel this order")
	ErrOrderStatusNotAllowed    = errors.New("order status not allowed")
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
	return status == constants.OrderStatusWaitAccept || status == constants.OrderStatusAccepted
}

// canCancelByOperator 校验取消方是否有权取消该订单。
func canCancelByOperator(order *model.RideOrder, operatorType string, operatorID int64) bool {
	switch operatorType {
	case constants.OperatorUser:
		return order.UserId == uint64(operatorID)
	case constants.OperatorDriver:
		return order.Status == constants.OrderStatusAccepted && order.DriverId == uint64(operatorID)
	case constants.OperatorSystem, constants.OperatorAdmin:
		return true
	default:
		return false
	}
}
