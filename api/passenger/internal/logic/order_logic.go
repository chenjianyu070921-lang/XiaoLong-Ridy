package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	priceclient "XiaoLong-Ridy/rpc/pricesvc/client"
)

// OrderLogic 封装乘客端订单相关业务流程。
type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewOrderLogic 创建订单业务逻辑实例。
func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *OrderLogic {
	return &OrderLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// CreateOrder 完成“预估价格 -> 创建订单”的乘客下单流程。
func (l *OrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if err := validateCreateOrder(req); err != nil {
		return nil, err
	}
	priceClient, err := l.priceClient()
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}

	price, err := priceClient.EstimatePrice(l.ctx, &priceclient.EstimatePriceRequest{
		CarType:       req.CarType,
		FromLongitude: req.FromLongitude,
		FromLatitude:  req.FromLatitude,
		ToLongitude:   req.ToLongitude,
		ToLatitude:    req.ToLatitude,
	})
	if err != nil {
		return nil, err
	}

	order, err := orderClient.CreateOrder(l.ctx, &orderproto.CreateOrderRequest{
		UserId:              int64(userID),
		CarType:             req.CarType,
		FromAddress:         strings.TrimSpace(req.FromAddress),
		FromLongitude:       req.FromLongitude,
		FromLatitude:        req.FromLatitude,
		ToAddress:           strings.TrimSpace(req.ToAddress),
		ToLongitude:         req.ToLongitude,
		ToLatitude:          req.ToLatitude,
		EstimatedDistanceM:  price.EstimatedDistanceM,
		EstimatedDurationS:  price.EstimatedDurationS,
		EstimatedPriceCents: price.EstimatedPriceCents,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateOrderResponse{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		EstimatedPriceCents: order.GetEstimatedPriceCents(),
		Status:              int32(order.GetStatus()),
		CreatedAt:           order.GetCreatedAt(),
	}, nil
}

// ListOrders 查询当前乘客自己的订单列表。
func (l *OrderLogic) ListOrders(req *types.ListOrdersRequest) (*types.ListOrdersResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := orderClient.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		UserId:   int64(userID),
		Status:   orderproto.OrderStatus(req.Status),
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.OrderSummary, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, types.OrderSummary{
			OrderID:             item.GetOrderId(),
			OrderNo:             item.GetOrderNo(),
			FromAddress:         item.GetFromAddress(),
			ToAddress:           item.GetToAddress(),
			Status:              int32(item.GetStatus()),
			EstimatedPriceCents: item.GetEstimatedPriceCents(),
			CreatedAt:           item.GetCreatedAt(),
		})
	}
	return &types.ListOrdersResponse{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

// GetOrder 查询当前乘客自己的订单详情。
func (l *OrderLogic) GetOrder(req *types.GetOrderRequest) (*types.OrderDetail, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	return toOrderDetail(order), nil
}

// CancelOrder 取消当前乘客自己的订单。
func (l *OrderLogic) CancelOrder(req *types.CancelOrderRequest) (*types.CancelOrderResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req.OrderID <= 0 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	current, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if current.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	resp, err := orderClient.CancelOrder(l.ctx, &orderproto.CancelOrderRequest{
		OrderId:      req.OrderID,
		OperatorType: "user",
		OperatorId:   int64(userID),
		Reason:       strings.TrimSpace(req.Reason),
	})
	if err != nil {
		return nil, err
	}
	return &types.CancelOrderResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

// validateCreateOrder 校验下单请求中的必填地址和车型参数。
func validateCreateOrder(req *types.CreateOrderRequest) error {
	if req == nil || strings.TrimSpace(req.FromAddress) == "" || strings.TrimSpace(req.ToAddress) == "" {
		return ErrInvalidRequest
	}
	if req.CarType <= 0 {
		return ErrInvalidRequest
	}
	return nil
}

// orderClient 获取订单服务客户端，避免业务方法重复判断空依赖。
func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}

// priceClient 获取价格服务客户端，供下单前预估价格使用。
func (l *OrderLogic) priceClient() (svc.PriceClient, error) {
	if l.svcCtx == nil || l.svcCtx.PriceClient == nil {
		return nil, ErrPriceClientNotConfigured
	}
	return l.svcCtx.PriceClient, nil
}

// toOrderDetail 将 ordersvc 的订单详情响应转换为乘客端 API 响应结构。
func toOrderDetail(order *orderproto.GetOrderResponse) *types.OrderDetail {
	return &types.OrderDetail{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		UserID:              order.GetUserId(),
		DriverID:            order.GetDriverId(),
		CarType:             order.GetCarType(),
		FromAddress:         order.GetFromAddress(),
		FromLongitude:       order.GetFromLongitude(),
		FromLatitude:        order.GetFromLatitude(),
		ToAddress:           order.GetToAddress(),
		ToLongitude:         order.GetToLongitude(),
		ToLatitude:          order.GetToLatitude(),
		EstimatedDistanceM:  order.GetEstimatedDistanceM(),
		EstimatedDurationS:  order.GetEstimatedDurationS(),
		EstimatedPriceCents: order.GetEstimatedPriceCents(),
		Status:              int32(order.GetStatus()),
		CancelReason:        order.GetCancelReason(),
		CancelBy:            order.GetCancelBy(),
		CreatedAt:           order.GetCreatedAt(),
		UpdatedAt:           order.GetUpdatedAt(),
	}
}
