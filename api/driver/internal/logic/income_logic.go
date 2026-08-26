package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

type IncomeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncomeLogic {
	return &IncomeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *IncomeLogic) GetIncomeSummary(driverID int64) (*types.GetIncomeSummaryResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	var totalIncome int64
	var completedOrders int64
	page := int32(1)
	pageSize := int32(100)
	for {
		resp, err := client.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
			DriverId: driverID,
			Status:   orderproto.OrderStatus_ORDER_STATUS_COMPLETED,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}
		for _, order := range resp.GetList() {
			totalIncome += order.GetEstimatedPriceCents()
			completedOrders++
		}
		if int64(page)*int64(pageSize) >= resp.GetTotal() || len(resp.GetList()) == 0 {
			break
		}
		page++
	}
	return &types.GetIncomeSummaryResponse{
		DriverID:         driverID,
		CompletedOrders:  completedOrders,
		TotalIncomeCents: totalIncome,
		Source:           "ordersvc.completed_orders",
	}, nil
}

func (l *IncomeLogic) ListIncomeBills(driverID int64, req *types.ListIncomeBillsRequest) (*types.ListIncomeBillsResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	page, pageSize := clampPage(req.Page, req.PageSize)
	resp, err := client.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus_ORDER_STATUS_COMPLETED,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	result := &types.ListIncomeBillsResponse{
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
		Source:   "ordersvc.completed_orders",
	}
	for _, order := range resp.GetList() {
		result.List = append(result.List, types.IncomeBill{
			OrderID:     order.GetOrderId(),
			OrderNo:     order.GetOrderNo(),
			IncomeCents: order.GetEstimatedPriceCents(),
			Status:      int32(order.GetStatus()),
			CreatedAt:   order.GetCreatedAt(),
		})
	}
	return result, nil
}

func (l *IncomeLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}
