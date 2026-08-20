package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// OrderLogic 封装司机接单与行程相关业务逻辑。
type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderLogic 构造订单业务逻辑处理器。
func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{ctx: ctx, svcCtx: svcCtx}
}

// AcceptOrder 当前登录司机接单。driverID 由鉴权中间件从 JWT 解析得到，orderID 来自请求体。
func (l *OrderLogic) AcceptOrder(driverID, orderID int64) (*types.AcceptOrderResponse, error) {
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.AcceptOrder(l.ctx, &orderproto.AcceptOrderRequest{
		OrderId:   orderID,
		DriverId:  driverID,
	})
	if err != nil {
		return nil, err
	}
	return &types.AcceptOrderResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// StartTrip 当前登录司机开始行程。
func (l *OrderLogic) StartTrip(driverID, orderID int64) (*types.StartTripResponse, error) {
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

// ConfirmArrive 当前登录司机确认已到达乘客上车点。
func (l *OrderLogic) ConfirmArrive(driverID, orderID int64) (*types.ConfirmArriveResponse, error) {
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

// FinishTrip 当前登录司机结束行程，并上报实际里程/时长/金额。
func (l *OrderLogic) FinishTrip(driverID int64, req *types.FinishTripRequest) (*types.FinishTripResponse, error) {
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

// orderClient 从服务上下文中安全取出 ordersvc 客户端。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}
