package logic

import (
	"context"
	"math"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetOrderLogic 创建订单详情逻辑对象。
func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrder 查询订单详情并映射为 RPC 响应。
func (l *GetOrderLogic) GetOrder(in *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}

	order, err := l.svcCtx.OrderRepository.GetByID(l.ctx, uint64(in.OrderId))
	if err != nil {
		return nil, err
	}

	return &proto.GetOrderResponse{
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
		EstimatedPriceCents: yuanToCents(order.EstimatedPrice),
		Status:              proto.OrderStatus(order.Status),
		CancelReason:        order.CancelReason,
		CancelBy:            order.CancelBy,
		CreatedAt:           order.CreatedAt.Unix(),
		UpdatedAt:           order.UpdatedAt.Unix(),
		CouponId:            order.CouponId,
		DiscountCents:       order.DiscountCents,
		PayableCents:        order.PayableCents,
		PaidCents:           order.PaidCents,
		RefundCents:         order.RefundCents,
		CityCode:            order.CityCode,
	}, nil
}

// yuanToCents 将元转成分。
func yuanToCents(yuan float64) int64 {
	return int64(math.Round(yuan * 100))
}
