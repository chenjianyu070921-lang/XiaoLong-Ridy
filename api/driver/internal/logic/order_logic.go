package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
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
	if driverID <= 0 || orderID <= 0 {
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

// StartTrip 当前登录司机开始行程。
func (l *OrderLogic) StartTrip(driverID, orderID int64) (*types.StartTripResponse, error) {
	if driverID <= 0 || orderID <= 0 {
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
	if driverClient, err := l.driverClient(); err == nil {
		_, err = driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
			DriverId:     driverID,
			OnlineStatus: 2,
		})
		if err != nil {
			return nil, err
		}
	}
	return &types.StartTripResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// ConfirmArrive 当前登录司机确认已到达乘客上车点。
func (l *OrderLogic) ConfirmArrive(driverID, orderID int64) (*types.ConfirmArriveResponse, error) {
	if driverID <= 0 || orderID <= 0 {
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

// FinishTrip 当前登录司机结束行程，并上报实际里程/时长/金额。
func (l *OrderLogic) FinishTrip(driverID int64, req *types.FinishTripRequest) (*types.FinishTripResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 || req.ActualDistanceM < 0 || req.ActualDurationS < 0 || req.ActualPriceCents < 0 {
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
	if driverClient, err := l.driverClient(); err == nil {
		_, err = driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
			DriverId:     driverID,
			OnlineStatus: 1,
		})
		if err != nil {
			return nil, err
		}
	}
	return &types.FinishTripResponse{
		OrderID:            resp.GetOrderId(),
		Status:             int32(resp.GetStatus()),
		PayableAmountCents: resp.GetPayableAmountCents(),
	}, nil
}

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
	return &types.RejectOrderResponse{
		OrderID:  resp.GetOrderId(),
		DriverID: resp.GetDriverId(),
		Status:   resp.GetStatus(),
	}, nil
}

func (l *OrderLogic) ListMyDispatches(driverID int64, page, pageSize, status int32) (*types.ListMyDispatchesResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPage(page, pageSize)
	resp, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		DriverId: driverID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.MyDispatchItem, 0, len(resp.GetList()))
	for _, record := range resp.GetList() {
		item := types.MyDispatchItem{
			Dispatch: types.DispatchRecord{
				ID:           record.GetId(),
				OrderID:      record.GetOrderId(),
				DriverID:     record.GetDriverId(),
				DispatchType: record.GetDispatchType(),
				Status:       record.GetStatus(),
				MatchScore:   record.GetMatchScore(),
				Remark:       record.GetRemark(),
				CreatedAt:    record.GetCreatedAt(),
				UpdatedAt:    record.GetUpdatedAt(),
			},
		}
		order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: record.GetOrderId()})
		if err != nil {
			return nil, err
		}
		item.Order = types.OrderBrief{
			OrderID:             order.GetOrderId(),
			OrderNo:             order.GetOrderNo(),
			FromAddress:         order.GetFromAddress(),
			ToAddress:           order.GetToAddress(),
			Status:              int32(order.GetStatus()),
			EstimatedPriceCents: order.GetEstimatedPriceCents(),
			CreatedAt:           order.GetCreatedAt(),
		}
		items = append(items, item)
	}
	return &types.ListMyDispatchesResponse{
		List:     items,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

func (l *OrderLogic) ListMyOrders(driverID int64, page, pageSize, status int32) (*types.ListMyOrdersResponse, error) {
	if driverID <= 0 || status < 0 || status > int32(orderproto.OrderStatus_ORDER_STATUS_CANCELLED) {
		return nil, ErrInvalidParam
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPage(page, pageSize)
	resp, err := orderClient.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus(status),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.OrderBrief, 0, len(resp.GetList()))
	for _, order := range resp.GetList() {
		items = append(items, types.OrderBrief{
			OrderID:             order.GetOrderId(),
			OrderNo:             order.GetOrderNo(),
			FromAddress:         order.GetFromAddress(),
			ToAddress:           order.GetToAddress(),
			Status:              int32(order.GetStatus()),
			EstimatedPriceCents: order.GetEstimatedPriceCents(),
			CreatedAt:           order.GetCreatedAt(),
		})
	}
	return &types.ListMyOrdersResponse{
		List:     items,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

// orderClient 从服务上下文中安全取出 ordersvc 客户端。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}

func (l *OrderLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

func (l *OrderLogic) dispatchClient() (svc.DispatchClient, error) {
	if l.svcCtx == nil || l.svcCtx.DispatchClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.DispatchClient, nil
}
