package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// OrderLogic 封装司机接单与行程相关业务逻辑。
// 它不直接访问数据库，而是把请求转发给下游 ordersvc（通过 ServiceContext.OrderClient），并负责状态推进与结果透传。
type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderLogic 构造订单业务逻辑处理器，注入请求上下文与服务上下文。
func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{ctx: ctx, svcCtx: svcCtx}
}

// AcceptOrder 司机接单。
// driverID 由鉴权中间件从 JWT 解析后传入（代表当前登录司机），orderID 来自请求体。
// 内部调用 ordersvc.AcceptOrder，将订单状态由「待接单(1)」推进到「已接单(2)」并绑定司机。
func (l *OrderLogic) AcceptOrder(driverID, orderID int64) (*types.AcceptOrderResponse, error) {
	if !validDriverOrder(driverID, orderID) {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.AcceptOrder(l.ctx, &orderproto.AcceptOrderRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	return &types.AcceptOrderResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// StartTrip 司机开始行程。
// driverID 来自 JWT，orderID 来自请求体。
// 内部调用 ordersvc.StartTrip，将订单状态由「已接单(2)」推进到「行程中(3)」。
func (l *OrderLogic) StartTrip(driverID, orderID int64) (*types.StartTripResponse, error) {
	if !validDriverOrder(driverID, orderID) {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.StartTrip(l.ctx, &orderproto.StartTripRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	return &types.StartTripResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// ConfirmArrive 司机确认已到达乘客上车点。
// driverID 来自 JWT，orderID 来自请求体。
// 内部调用 ordersvc.ConfirmArrive，写入「司机已到达」状态日志（订单状态保持不变，仍为「已接单(2)」），
// 用于告知后续开始行程的前置节点。
func (l *OrderLogic) ConfirmArrive(driverID, orderID int64) (*types.ConfirmArriveResponse, error) {
	if !validDriverOrder(driverID, orderID) {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ConfirmArrive(l.ctx, &orderproto.ConfirmArriveRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	return &types.ConfirmArriveResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// FinishTrip 司机结束行程，并上报实际里程/时长/金额。
// driverID 来自 JWT；req 含 orderID 与实际上报数据（里程/时长/金额）。
// 内部调用 ordersvc.FinishTrip，将订单状态由「行程中(3)」推进到「待支付(4)」，并返回应付金额。
func (l *OrderLogic) FinishTrip(driverID int64, req *types.FinishTripRequest) (*types.FinishTripResponse, error) {
	if driverID <= 0 || !validFinishTripRequest(req) {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.FinishTrip(l.ctx, &orderproto.FinishTripRequest{
		OrderId:          req.OrderID,
		DriverId:         driverID,
		ActualDistanceM:  req.ActualDistanceM,
		ActualDurationS:  req.ActualDurationS,
		ActualPriceCents: req.ActualPriceCents,
	})
	if err != nil {
		return nil, err
	}
	return &types.FinishTripResponse{
		OrderID:            resp.GetOrderId(),
		Status:             int32(resp.GetStatus()),
		PayableAmountCents: resp.GetPayableAmountCents(),
	}, nil
}

// RejectOrder 司机拒绝派单。
// TODO: 待 dispatchsvc 提供 RejectDispatch 后，在这里调用该 RPC；当前只保留司机端业务骨架，不实际调用下游。
func (l *OrderLogic) RejectOrder(driverID int64, req *types.RejectOrderRequest) (*types.RejectOrderResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidParam
	}
	return nil, ErrDispatchDependencyNotReady
}

// ListMyDispatches 查询当前司机的派单记录。
// TODO: 待 dispatchsvc 提供按 driver_id 查询能力后，在这里调用 ListDispatchRecords；当前只保留司机端业务骨架，不实际调用下游。
func (l *OrderLogic) ListMyDispatches(driverID int64, page, pageSize int32) (*types.ListMyDispatchesResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	page, pageSize = clampPage(page, pageSize)
	return nil, ErrDispatchDependencyNotReady
}

// orderClient 从服务上下文中安全取出 ordersvc 客户端。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}
