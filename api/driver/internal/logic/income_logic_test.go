package logic

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
)

func TestGetIncomeSummaryUsesPaysvcSettlements(t *testing.T) {
	client := &fakePayClient{listSettlementsResponse: &payproto.ListSettlementsResponse{
		Records: []*payproto.SettlementBill{
			{SettlementId: 1, SettlementNo: "SET1001", OrderId: 1001, DriverId: 25, DriverIncomeCents: 1200, Status: 2, SettledAt: 100},
			{SettlementId: 2, SettlementNo: "SET1002", OrderId: 1002, DriverId: 25, DriverIncomeCents: 2300, Status: 2, SettledAt: 200},
		},
		Total:    2,
		Page:     1,
		PageSize: 100,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{PayClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.DriverID != 25 || resp.CompletedOrders != 2 || resp.TotalIncomeCents != 3500 || resp.Source != "paysvc.settlement" {
		t.Fatalf("GetIncomeSummary() response = %+v", resp)
	}
	if resp.WithdrawableCents != 3500 {
		t.Fatalf("GetIncomeSummary() withdrawable cents = %d, want 3500", resp.WithdrawableCents)
	}
	if client.listSettlementsRequest.GetDriverId() != 25 ||
		client.listSettlementsRequest.GetStatus() != 2 ||
		client.listSettlementsRequest.GetPageSize() != 100 {
		t.Fatalf("ListSettlements() request = %+v", client.listSettlementsRequest)
	}
}

func TestListIncomeBillsReturnsSettlementBills(t *testing.T) {
	client := &fakePayClient{listSettlementsResponse: &payproto.ListSettlementsResponse{
		Records: []*payproto.SettlementBill{
			{SettlementId: 8, SettlementNo: "SET1001", OrderId: 1001, DriverId: 25, DriverIncomeCents: 1200, Status: 2, SettledAt: 123},
		},
		Total:    1,
		Page:     2,
		PageSize: 8,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{PayClient: client})

	resp, err := logic.ListIncomeBills(25, &types.ListIncomeBillsRequest{Page: 2, PageSize: 8})
	if err != nil {
		t.Fatalf("ListIncomeBills() error = %v", err)
	}
	if resp.Total != 1 || resp.Page != 2 || resp.PageSize != 8 || resp.Source != "paysvc.settlement" || len(resp.List) != 1 {
		t.Fatalf("ListIncomeBills() response = %+v", resp)
	}
	if resp.List[0].OrderID != 1001 || resp.List[0].OrderNo != "SET1001" ||
		resp.List[0].IncomeCents != 1200 || resp.List[0].Status != 2 || resp.List[0].CreatedAt != 123 {
		t.Fatalf("income bill = %+v", resp.List[0])
	}
}

func TestGetTodayIncomeFiltersPaysvcSettlementsForCurrentDay(t *testing.T) {
	originalNow := incomeNow
	incomeNow = func() time.Time { return time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC) }
	defer func() { incomeNow = originalNow }()

	start := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC).Unix()
	client := &fakePayClient{listSettlementsResponse: &payproto.ListSettlementsResponse{
		Records: []*payproto.SettlementBill{
			{SettlementId: 1, SettlementNo: "SET2001", OrderId: 2001, DriverId: 25, DriverIncomeCents: 1200, Status: 2, SettledAt: start + 60},
		},
		Total:    1,
		Page:     1,
		PageSize: 100,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{PayClient: client})

	resp, err := logic.GetTodayIncome(25)
	if err != nil {
		t.Fatalf("GetTodayIncome() error = %v", err)
	}
	if resp.Period != "today" || resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 1200 {
		t.Fatalf("GetTodayIncome() response = %+v", resp)
	}
	if client.listSettlementsRequest.GetStartAt() != start || client.listSettlementsRequest.GetEndAt() != start+86400 {
		t.Fatalf("ListSettlements() window = %d-%d, want %d-%d", client.listSettlementsRequest.GetStartAt(), client.listSettlementsRequest.GetEndAt(), start, start+86400)
	}
}

func TestGetWeekIncomeFiltersPaysvcSettlementsFromMonday(t *testing.T) {
	originalNow := incomeNow
	incomeNow = func() time.Time { return time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC) }
	defer func() { incomeNow = originalNow }()

	weekStart := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC).Unix()
	client := &fakePayClient{listSettlementsResponse: &payproto.ListSettlementsResponse{
		Records: []*payproto.SettlementBill{
			{SettlementId: 3, SettlementNo: "SET3001", OrderId: 3001, DriverId: 25, DriverIncomeCents: 2200, Status: 2, SettledAt: weekStart + 60},
		},
		Total:    1,
		Page:     1,
		PageSize: 100,
	}}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{PayClient: client})

	resp, err := logic.GetWeekIncome(25)
	if err != nil {
		t.Fatalf("GetWeekIncome() error = %v", err)
	}
	if resp.Period != "week" || resp.DriverID != 25 || resp.CompletedOrders != 1 || resp.TotalIncomeCents != 2200 {
		t.Fatalf("GetWeekIncome() response = %+v", resp)
	}
	if client.listSettlementsRequest.GetStartAt() != weekStart || client.listSettlementsRequest.GetEndAt() != weekStart+7*86400 {
		t.Fatalf("ListSettlements() window = %d-%d, want %d-%d", client.listSettlementsRequest.GetStartAt(), client.listSettlementsRequest.GetEndAt(), weekStart, weekStart+7*86400)
	}
}

func TestGetIncomeSummaryAggregatesMultipleSettlementPages(t *testing.T) {
	client := &fakePayClient{
		listSettlementsResponses: map[int32]*payproto.ListSettlementsResponse{
			1: &payproto.ListSettlementsResponse{
				Records: []*payproto.SettlementBill{
					{SettlementId: 1, SettlementNo: "SET1001", OrderId: 1001, DriverId: 25, DriverIncomeCents: 1200, Status: 2, SettledAt: 100},
					{SettlementId: 2, SettlementNo: "SET1002", OrderId: 1002, DriverId: 25, DriverIncomeCents: 2300, Status: 2, SettledAt: 200},
				},
				Total:    250,
				Page:     1,
				PageSize: 100,
			},
			2: &payproto.ListSettlementsResponse{
				Records: []*payproto.SettlementBill{
					{SettlementId: 3, SettlementNo: "SET1003", OrderId: 1003, DriverId: 25, DriverIncomeCents: 3400, Status: 2, SettledAt: 300},
				},
				Total:    250,
				Page:     2,
				PageSize: 100,
			},
			3: &payproto.ListSettlementsResponse{
				Records: []*payproto.SettlementBill{
					{SettlementId: 4, SettlementNo: "SET1004", OrderId: 1004, DriverId: 25, DriverIncomeCents: 4500, Status: 2, SettledAt: 400},
				},
				Total:    250,
				Page:     3,
				PageSize: 100,
			},
		},
	}
	logic := NewIncomeLogic(context.Background(), &svc.ServiceContext{PayClient: client})

	resp, err := logic.GetIncomeSummary(25)
	if err != nil {
		t.Fatalf("GetIncomeSummary() error = %v", err)
	}
	if resp.CompletedOrders != 4 || resp.TotalIncomeCents != 11400 {
		t.Fatalf("GetIncomeSummary() response = %+v", resp)
	}

	pages := client.requestPages()
	sort.Slice(pages, func(i, j int) bool { return pages[i] < pages[j] })
	if len(pages) != 3 || pages[0] != 1 || pages[1] != 2 || pages[2] != 3 {
		t.Fatalf("ListSettlements() pages = %v, want [1 2 3]", pages)
	}
}

type fakePayClient struct {
	mu                       sync.Mutex
	listSettlementsRequest   *payproto.ListSettlementsRequest
	listSettlementsRequests  []*payproto.ListSettlementsRequest
	listSettlementsResponse  *payproto.ListSettlementsResponse
	listSettlementsResponses map[int32]*payproto.ListSettlementsResponse
}

func (f *fakePayClient) ListSettlements(_ context.Context, req *payproto.ListSettlementsRequest) (*payproto.ListSettlementsResponse, error) {
	f.mu.Lock()
	f.listSettlementsRequest = req
	f.listSettlementsRequests = append(f.listSettlementsRequests, req)
	resp := f.listSettlementsResponse
	if f.listSettlementsResponses != nil {
		if pageResp, ok := f.listSettlementsResponses[req.GetPage()]; ok {
			resp = pageResp
		}
	}
	f.mu.Unlock()
	if resp != nil {
		return resp, nil
	}
	return &payproto.ListSettlementsResponse{Page: req.GetPage(), PageSize: req.GetPageSize()}, nil
}

func (f *fakePayClient) requestPages() []int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	pages := make([]int32, 0, len(f.listSettlementsRequests))
	for _, req := range f.listSettlementsRequests {
		pages = append(pages, req.GetPage())
	}
	return pages
}
