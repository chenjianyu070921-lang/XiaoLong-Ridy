package logic

import (
	"context"
	"testing"

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
	if resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 29900 {
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
	if resp.List[0].OrderID != 1001 || resp.List[0].IncomeCents != 29900 || resp.List[0].Status != int32(orderproto.OrderStatus_ORDER_STATUS_COMPLETED) {
		t.Fatalf("income bill = %+v", resp.List[0])
	}
}
