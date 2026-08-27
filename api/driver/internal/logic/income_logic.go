package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

var incomeNow = time.Now

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
			incomeCents, err := l.orderIncomeCents(client, order.GetOrderId())
			if err != nil {
				return nil, err
			}
			totalIncome += incomeCents
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

func (l *IncomeLogic) GetTodayIncome(driverID int64) (*types.PeriodIncomeResponse, error) {
	now := incomeNow()
	start := beginningOfDay(now)
	return l.getPeriodIncome(driverID, "today", start, start.AddDate(0, 0, 1))
}

func (l *IncomeLogic) GetWeekIncome(driverID int64) (*types.PeriodIncomeResponse, error) {
	now := incomeNow()
	start := beginningOfWeek(now)
	return l.getPeriodIncome(driverID, "week", start, start.AddDate(0, 0, 7))
}

func (l *IncomeLogic) getPeriodIncome(driverID int64, period string, start, end time.Time) (*types.PeriodIncomeResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	startAt := start.Unix()
	endAt := end.Unix()
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
			createdAt := order.GetCreatedAt()
			if createdAt >= startAt && createdAt < endAt {
				incomeCents, err := l.orderIncomeCents(client, order.GetOrderId())
				if err != nil {
					return nil, err
				}
				totalIncome += incomeCents
				completedOrders++
			}
		}
		if int64(page)*int64(pageSize) >= resp.GetTotal() || len(resp.GetList()) == 0 {
			break
		}
		page++
	}
	return &types.PeriodIncomeResponse{
		DriverID:         driverID,
		Period:           period,
		CompletedOrders:  completedOrders,
		TotalIncomeCents: totalIncome,
		StartAt:          startAt,
		EndAt:            endAt,
		Source:           "ordersvc.completed_orders",
	}, nil
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func beginningOfWeek(t time.Time) time.Time {
	dayStart := beginningOfDay(t)
	offset := (int(dayStart.Weekday()) + 6) % 7
	return dayStart.AddDate(0, 0, -offset)
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
		incomeCents, err := l.orderIncomeCents(client, order.GetOrderId())
		if err != nil {
			return nil, err
		}
		result.List = append(result.List, types.IncomeBill{
			OrderID:     order.GetOrderId(),
			OrderNo:     order.GetOrderNo(),
			IncomeCents: incomeCents,
			Status:      int32(order.GetStatus()),
			CreatedAt:   order.GetCreatedAt(),
		})
	}
	return result, nil
}

func (l *IncomeLogic) orderIncomeCents(client svc.OrderClient, orderID int64) (int64, error) {
	order, err := client.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return 0, err
	}
	if order == nil {
		return 0, nil
	}

	// Driver income follows the settled order amount, not the booking estimate.
	amount := order.GetPayableCents()
	if amount <= 0 {
		amount = order.GetPaidCents()
	}
	amount -= order.GetRefundCents()
	if amount < 0 {
		return 0, nil
	}
	return amount, nil
}

func (l *IncomeLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}
