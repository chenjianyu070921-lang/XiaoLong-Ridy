package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTimeoutOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTimeoutOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTimeoutOrdersLogic {
	return &ListTimeoutOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListTimeoutOrders 查询超过 timeout_seconds 仍未接单的订单，供超时扫描任务分页拉取。
func (l *ListTimeoutOrdersLogic) ListTimeoutOrders(in *proto.ListTimeoutOrdersRequest) (*proto.ListTimeoutOrdersResponse, error) {
	if in.TimeoutSeconds < 0 {
		return nil, ErrInvalidOrderParams
	}
	timeout := time.Duration(in.TimeoutSeconds) * time.Second
	if in.TimeoutSeconds == 0 {
		timeout = 5 * time.Minute
	}
	page := normalizePage(in.Page)
	pageSize := normalizePageSize(in.PageSize)
	before := time.Now().Add(-timeout)

	list, total, err := l.svcCtx.OrderRepository.ListTimeoutOrders(l.ctx, before, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*proto.OrderSummary, 0, len(list))
	for i := range list {
		order := list[i]
		items = append(items, &proto.OrderSummary{
			OrderId:             int64(order.Id),
			OrderNo:             order.OrderNo,
			FromAddress:         order.FromAddress,
			ToAddress:           order.ToAddress,
			Status:              proto.OrderStatus(order.Status),
			EstimatedPriceCents: yuanToCents(order.EstimatedPrice),
			CreatedAt:           order.CreatedAt.Unix(),
		})
	}

	return &proto.ListTimeoutOrdersResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
