package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
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
func (l *OrderLogic) RejectOrder(driverID int64, req *types.RejectOrderRequest) (*types.RejectOrderResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.RejectDispatch(l.ctx, &dispatchproto.RejectDispatchRequest{
		OrderId:  req.OrderID,
		DriverId: driverID,
		Reason:   req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return &types.RejectOrderResponse{OrderID: resp.GetOrderId(), DriverID: resp.GetDriverId(), Status: resp.GetStatus()}, nil
}

// ListMyDispatches 查询当前司机的派单记录。
func (l *OrderLogic) ListMyDispatches(driverID int64, page, pageSize, status int32) (*types.ListMyDispatchesResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	page, pageSize = clampPage(page, pageSize)
	if status < 0 {
		return nil, ErrInvalidParam
	}
	if status == 0 {
		status = 1
	}
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	dispatches, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		DriverId: driverID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.MyDispatchItem, 0, len(dispatches.GetList()))
	for _, dispatch := range dispatches.GetList() {
		order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: dispatch.GetOrderId()})
		if err != nil {
			return nil, err
		}
		items = append(items, types.MyDispatchItem{
			Dispatch: types.DispatchRecord{
				ID: dispatch.GetId(), OrderID: dispatch.GetOrderId(), DriverID: dispatch.GetDriverId(),
				DispatchType: dispatch.GetDispatchType(), Status: dispatch.GetStatus(), MatchScore: dispatch.GetMatchScore(),
				Remark: dispatch.GetRemark(), CreatedAt: dispatch.GetCreatedAt(), UpdatedAt: dispatch.GetUpdatedAt(),
			},
			Order: types.OrderBrief{
				OrderID: order.GetOrderId(), OrderNo: order.GetOrderNo(), FromAddress: order.GetFromAddress(),
				ToAddress: order.GetToAddress(), Status: int32(order.GetStatus()),
				EstimatedPriceCents: order.GetEstimatedPriceCents(), CreatedAt: order.GetCreatedAt(),
			},
		})
	}
	return &types.ListMyDispatchesResponse{List: items, Total: dispatches.GetTotal(), Page: dispatches.GetPage(), PageSize: dispatches.GetPageSize()}, nil
}

// orderClient 从服务上下文中安全取出 ordersvc 客户端。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}

func (l *OrderLogic) dispatchClient() (svc.DispatchClient, error) {
	if l.svcCtx == nil || l.svcCtx.DispatchClient == nil {
		return nil, ErrDispatchClientNotConfigured
	}
	return l.svcCtx.DispatchClient, nil
}
