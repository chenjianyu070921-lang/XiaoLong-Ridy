package client

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// LocalClient 是本地开发和测试使用的订单服务实现。
type LocalClient struct {
	mu     sync.RWMutex
	nextID uint64
	orders map[uint64]*orderproto.GetOrderResponse
}

// NewLocalClient 创建本地订单服务实现。
func NewLocalClient() *LocalClient {
	return &LocalClient{
		nextID: 1,
		orders: make(map[uint64]*orderproto.GetOrderResponse),
	}
}

// CreateOrder 创建订单并返回订单号与预估费用。
func (c *LocalClient) CreateOrder(_ context.Context, req *orderproto.CreateOrderRequest) (*orderproto.CreateOrderResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++
	now := time.Now().Unix()
	orderNo := fmt.Sprintf("ORD%014d", id)
	status := orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT
	order := &orderproto.GetOrderResponse{
		OrderId:             int64(id),
		OrderNo:             orderNo,
		UserId:              req.GetUserId(),
		DriverId:            0,
		CarType:             req.GetCarType(),
		FromAddress:         req.GetFromAddress(),
		FromLongitude:       req.GetFromLongitude(),
		FromLatitude:        req.GetFromLatitude(),
		ToAddress:           req.GetToAddress(),
		ToLongitude:         req.GetToLongitude(),
		ToLatitude:          req.GetToLatitude(),
		EstimatedDistanceM:  req.GetEstimatedDistanceM(),
		EstimatedDurationS:  req.GetEstimatedDurationS(),
		EstimatedPriceCents: req.GetEstimatedPriceCents(),
		Status:              status,
		CancelReason:        "",
		CancelBy:            "",
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	c.orders[id] = order

	return &orderproto.CreateOrderResponse{
		OrderId:             int64(id),
		OrderNo:             orderNo,
		EstimatedPriceCents: req.GetEstimatedPriceCents(),
		Status:              status,
		CreatedAt:           now,
	}, nil
}

// CancelOrder 取消订单并记录取消原因。
func (c *LocalClient) CancelOrder(_ context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	order, ok := c.orders[uint64(req.GetOrderId())]
	if !ok {
		return nil, fmt.Errorf("orderclient not found")
	}
	order.Status = orderproto.OrderStatus_ORDER_STATUS_CANCELLED
	order.CancelReason = req.GetReason()
	order.CancelBy = req.GetOperatorType()
	order.UpdatedAt = time.Now().Unix()
	return &orderproto.CancelOrderResponse{
		OrderId: order.GetOrderId(),
		Status:  order.Status,
	}, nil
}

// GetOrder 返回订单详情。
func (c *LocalClient) GetOrder(_ context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, ok := c.orders[uint64(req.GetOrderId())]
	if !ok {
		return nil, fmt.Errorf("orderclient not found")
	}
	return &orderproto.GetOrderResponse{
		OrderId:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		UserId:              order.GetUserId(),
		DriverId:            order.GetDriverId(),
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
		Status:              order.GetStatus(),
		CancelReason:        order.GetCancelReason(),
		CancelBy:            order.GetCancelBy(),
		CreatedAt:           order.GetCreatedAt(),
		UpdatedAt:           order.GetUpdatedAt(),
	}, nil
}

// ListOrders 返回订单分页列表。
func (c *LocalClient) ListOrders(_ context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]*orderproto.OrderSummary, 0, len(c.orders))
	for _, order := range c.orders {
		if req.GetUserId() > 0 && order.GetUserId() != req.GetUserId() {
			continue
		}
		if req.GetDriverId() > 0 && order.GetDriverId() != req.GetDriverId() {
			continue
		}
		if req.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_UNSPECIFIED && order.GetStatus() != req.GetStatus() {
			continue
		}
		list = append(list, &orderproto.OrderSummary{
			OrderId:             order.GetOrderId(),
			OrderNo:             order.GetOrderNo(),
			FromAddress:         order.GetFromAddress(),
			ToAddress:           order.GetToAddress(),
			Status:              order.GetStatus(),
			EstimatedPriceCents: order.GetEstimatedPriceCents(),
			CreatedAt:           order.GetCreatedAt(),
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt > list[j].CreatedAt
	})

	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	start := int((page - 1) * pageSize)
	if start > len(list) {
		start = len(list)
	}
	end := start + int(pageSize)
	if end > len(list) {
		end = len(list)
	}

	return &orderproto.ListOrdersResponse{
		List:     list[start:end],
		Total:    int64(len(list)),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
