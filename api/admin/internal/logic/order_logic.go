package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// OrderLogic 封装管理后台订单监控业务。
// P0 阶段提供列表和详情聚合能力，方便客服和运营排查订单。
type OrderLogic struct {
	ctx *svc.ServiceContext
}

// NewOrderLogic 创建订单业务逻辑对象。
func NewOrderLogic(ctx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{ctx: ctx}
}

// List 查询订单列表。
func (l *OrderLogic) List(ctx context.Context, req types.OrderListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.OrderRepository.List(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]types.OrderDTO, 0, len(list))
	for _, item := range list {
		items = append(items, toOrderDTO(item))
	}
	return &types.PageResult{
		List:     items,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// ListAbnormal 查询异常订单列表。
// 当前实现只读取订单、支付和派单状态，不执行退款、改派等高风险写操作。
func (l *OrderLogic) ListAbnormal(ctx context.Context, req types.AbnormalOrderListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.OrderRepository.ListAbnormal(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.PageResult{
		List:     list,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// Detail 查询订单详情并聚合关联数据。
func (l *OrderLogic) Detail(ctx context.Context, id int64) (*types.OrderDetailDTO, error) {
	order, err := l.ctx.OrderRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	statusLogs, err := l.ctx.OrderRepository.ListStatusLogs(ctx, id)
	if err != nil {
		return nil, err
	}
	dispatchRecords, err := l.ctx.OrderRepository.ListDispatchRecords(ctx, id)
	if err != nil {
		return nil, err
	}
	price, err := l.ctx.OrderRepository.GetOrderPrice(ctx, id)
	if err != nil {
		return nil, err
	}
	payment, err := l.ctx.OrderRepository.GetPayment(ctx, id)
	if err != nil {
		return nil, err
	}
	settlement, err := l.ctx.OrderRepository.GetSettlement(ctx, id)
	if err != nil {
		return nil, err
	}
	return &types.OrderDetailDTO{
		Order:           toOrderDTO(*order),
		StatusLogs:      statusLogs,
		DispatchRecords: dispatchRecords,
		Price:           price,
		Payment:         payment,
		Settlement:      settlement,
	}, nil
}

// toOrderDTO 将订单数据库模型转换为接口 DTO。
func toOrderDTO(order model.RideOrder) types.OrderDTO {
	return types.OrderDTO{
		ID:                 order.ID,
		OrderNo:            order.OrderNo,
		UserID:             order.UserID,
		DriverID:           order.DriverID,
		CarType:            order.CarType,
		FromAddress:        order.FromAddress,
		FromLongitude:      order.FromLongitude,
		FromLatitude:       order.FromLatitude,
		ToAddress:          order.ToAddress,
		ToLongitude:        order.ToLongitude,
		ToLatitude:         order.ToLatitude,
		EstimatedDistanceM: order.EstimatedDistanceM,
		EstimatedDurationS: order.EstimatedDurationS,
		EstimatedPrice:     order.EstimatedPrice,
		Status:             order.Status,
		CancelReason:       order.CancelReason,
		CancelBy:           order.CancelBy,
		CreatedAt:          repository.FormatTime(order.CreatedAt),
		UpdatedAt:          repository.FormatTime(order.UpdatedAt),
	}
}
