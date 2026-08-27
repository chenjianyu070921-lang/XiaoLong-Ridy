package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestGetIncomeSummaryDerivesCompletedOrderIncome(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 1001 {
		t.Fatalf("GetIncomeSummary() response = %+v", resp)
	}
	if client.listOrdersRequest.GetDriverId() != 25 ||
		client.listOrdersRequest.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_COMPLETED ||
		client.listOrdersRequest.GetPageSize() != 100 {
		t.Fatalf("ListOrders() request = %+v", client.listOrdersRequest)
	}
}

func TestListIncomeBillsReturnsCompletedOrderBills(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.ListIncomeBills(25, &types.ListIncomeBillsRequest{Page: 2, PageSize: 8})
	if err != nil {
		t.Fatalf("ListIncomeBills() error = %v", err)
	}
	if resp.Total != 1 || resp.Page != 2 || resp.PageSize != 8 || len(resp.List) != 1 {
		t.Fatalf("ListIncomeBills() response = %+v", resp)
	}
	if resp.List[0].OrderID != 1001 || resp.List[0].IncomeCents != 1001 || resp.List[0].Status != int32(orderproto.OrderStatus_ORDER_STATUS_COMPLETED) {
		t.Fatalf("income bill = %+v", resp.List[0])
	}
}

func TestIncomeUsesSettledAmountAndRefundInsteadOfEstimate(t *testing.T) {
	client := &fakeOrderClient{getOrderResponses: map[int64]*orderproto.GetOrderResponse{
		1001: {OrderId: 1001, EstimatedPriceCents: 29900, PayableCents: 1500, RefundCents: 300},
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.TotalIncomeCents != 1200 {
		t.Fatalf("GetIncomeSummary() income = %d, want settled amount 1200", resp.TotalIncomeCents)
	}
}

func TestIncomeFallsBackToPaidCentsWhenPayableMissing(t *testing.T) {
	client := &fakeOrderClient{getOrderResponses: map[int64]*orderproto.GetOrderResponse{
		1001: {OrderId: 1001, EstimatedPriceCents: 29900, PaidCents: 1800},
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.TotalIncomeCents != 1800 {
		t.Fatalf("GetIncomeSummary() income = %d, want paid amount 1800", resp.TotalIncomeCents)
	}
}

func TestIncomeDoesNotReturnNegativeWhenRefundExceedsAmount(t *testing.T) {
	client := &fakeOrderClient{getOrderResponses: map[int64]*orderproto.GetOrderResponse{
		1001: {OrderId: 1001, PayableCents: 1000, RefundCents: 1200},
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.TotalIncomeCents != 0 {
		t.Fatalf("GetIncomeSummary() income = %d, want 0", resp.TotalIncomeCents)
	}
}

func TestGetTodayIncomeFiltersCompletedOrdersForCurrentDay(t *testing.T) {
	originalNow := incomeNow
	incomeNow = func() time.Time { return time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC) }
	defer func() { incomeNow = originalNow }()

	start := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC).Unix()
	client := &fakeOrderClient{listOrdersResponse: &orderproto.ListOrdersResponse{
		List: []*orderproto.OrderSummary{
			{OrderId: 2001, OrderNo: "today", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 1200, CreatedAt: start + 60},
			{OrderId: 2002, OrderNo: "yesterday", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 3400, CreatedAt: start - 1},
			{OrderId: 2003, OrderNo: "tomorrow", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 5600, CreatedAt: start + 86400},
		},
		Total:    3,
		Page:     1,
		PageSize: 100,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetTodayIncome(25)
	if err != nil {
		t.Fatalf("GetTodayIncome() error = %v", err)
	}
	if resp.Period != "today" || resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 1001 {
		t.Fatalf("GetTodayIncome() response = %+v", resp)
	}
	if resp.StartAt != start || resp.EndAt != start+86400 {
		t.Fatalf("GetTodayIncome() window = %d-%d, want %d-%d", resp.StartAt, resp.EndAt, start, start+86400)
	}
}

func TestGetWeekIncomeFiltersCompletedOrdersFromMonday(t *testing.T) {
	originalNow := incomeNow
	incomeNow = func() time.Time { return time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC) }
	defer func() { incomeNow = originalNow }()

	weekStart := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC).Unix()
	client := &fakeOrderClient{listOrdersResponse: &orderproto.ListOrdersResponse{
		List: []*orderproto.OrderSummary{
			{OrderId: 3001, OrderNo: "monday", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 2200, CreatedAt: weekStart + 60},
			{OrderId: 3002, OrderNo: "sunday-before", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 3300, CreatedAt: weekStart - 1},
			{OrderId: 3003, OrderNo: "next-week", Status: orderproto.OrderStatus_ORDER_STATUS_COMPLETED, EstimatedPriceCents: 4400, CreatedAt: weekStart + 7*86400},
		},
		Total:    3,
		Page:     1,
		PageSize: 100,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetWeekIncome(25)
	if err != nil {
		t.Fatalf("GetWeekIncome() error = %v", err)
	}
	if resp.Period != "week" || resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 1001 {
		t.Fatalf("GetWeekIncome() response = %+v", resp)
	}
	if resp.StartAt != weekStart || resp.EndAt != weekStart+7*86400 {
		t.Fatalf("GetWeekIncome() window = %d-%d, want %d-%d", resp.StartAt, resp.EndAt, weekStart, weekStart+7*86400)
	}
}
