package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListOrdersLogic 创建订单列表逻辑对象。
func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrders 按用户/司机/状态分页查询订单摘要。
func (l *ListOrdersLogic) ListOrders(in *proto.ListOrdersRequest) (*proto.ListOrdersResponse, error) {
	// 上界取 REFUNDED(7) 而非 CANCELLED(6)，否则已退款订单无法被筛选（对账/后台查不到）。
	if in.Status < 0 || in.Status > proto.OrderStatus_ORDER_STATUS_REFUNDED {
		return nil, ErrInvalidOrderParams
	}
	page := normalizePage(in.Page)
	pageSize := normalizePageSize(in.PageSize)

	list, total, err := l.svcCtx.OrderRepository.List(l.ctx, uint64(in.UserId), uint64(in.DriverId), int8(in.Status), page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*proto.OrderSummary, 0, len(list))
	for i := range list {
		order := list[i]
		items = append(items, &proto.OrderSummary{
			OrderId:             int64(order.Id),
			OrderNo:             order.OrderNo,
			UserId:              int64(order.UserId),
			DriverId:            int64(order.DriverId),
			CarType:             int32(order.CarType),
			FromAddress:         order.FromAddress,
			FromLongitude:       order.FromLongitude,
			FromLatitude:        order.FromLatitude,
			ToAddress:           order.ToAddress,
			ToLongitude:         order.ToLongitude,
			ToLatitude:          order.ToLatitude,
			EstimatedDistanceM:  int64(order.EstimatedDistanceM),
			EstimatedDurationS:  int64(order.EstimatedDurationS),
			Status:              proto.OrderStatus(order.Status),
			EstimatedPriceCents: yuanToCents(order.EstimatedPrice),
			CancelReason:        order.CancelReason,
			CancelBy:            order.CancelBy,
			CreatedAt:           order.CreatedAt.Unix(),
			UpdatedAt:           order.UpdatedAt.Unix(),
		})
	}

	return &proto.ListOrdersResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// normalizePage 归一化页码，默认 1。
func normalizePage(page int32) int32 {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizePageSize 归一化每页条数，默认 20，上限 100。
func normalizePageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
