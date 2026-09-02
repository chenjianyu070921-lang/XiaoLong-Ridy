package logic

import (
	"context"
	"sync"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
)

const (
	incomePageSize          int32 = 100
	incomePageFetchWorkers        = 8
	maxIncomeSettlementPages      = 20
	incomeSource                  = "paysvc.settlement"
	settlementStatusSettled int32 = 2
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
	summary, err := l.sumSettlements(driverID, 0, 0)
	if err != nil {
		return nil, err
	}
	source := incomeSource
	if summary.capped {
		source = incomeSource + ":capped"
	}
	return &types.GetIncomeSummaryResponse{
		DriverID:         driverID,
		CompletedOrders:  summary.count,
		TotalIncomeCents: summary.incomeCents,
		Source:           source,
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
	startAt := start.Unix()
	endAt := end.Unix()
	summary, err := l.sumSettlements(driverID, startAt, endAt)
	if err != nil {
		return nil, err
	}
	source := incomeSource
	if summary.capped {
		source = incomeSource + ":capped"
	}
	return &types.PeriodIncomeResponse{
		DriverID:         driverID,
		Period:           period,
		CompletedOrders:  summary.count,
		TotalIncomeCents: summary.incomeCents,
		StartAt:          startAt,
		EndAt:            endAt,
		Source:           source,
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
	client, err := l.payClient()
	if err != nil {
		return nil, err
	}
	page, pageSize := clampPage(req.Page, req.PageSize)
	resp, err := client.ListSettlements(l.ctx, &payproto.ListSettlementsRequest{
		DriverId: driverID,
		Status:   settlementStatusSettled,
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
		Source:   incomeSource,
	}
	for _, bill := range resp.GetRecords() {
		result.List = append(result.List, types.IncomeBill{
			OrderID:     bill.GetOrderId(),
			OrderNo:     bill.GetSettlementNo(),
			IncomeCents: bill.GetDriverIncomeCents(),
			Status:      bill.GetStatus(),
			CreatedAt:   settlementDisplayTime(bill),
		})
	}
	return result, nil
}

type incomeSettlementSummary struct {
	count       int64
	incomeCents int64
	capped      bool
}

func (l *IncomeLogic) sumSettlements(driverID, startAt, endAt int64) (incomeSettlementSummary, error) {
	client, err := l.payClient()
	if err != nil {
		return incomeSettlementSummary{}, err
	}
	firstResp, err := client.ListSettlements(l.ctx, &payproto.ListSettlementsRequest{
		DriverId: driverID,
		Status:   settlementStatusSettled,
		StartAt:  startAt,
		EndAt:    endAt,
		Page:     1,
		PageSize: incomePageSize,
	})
	if err != nil {
		return incomeSettlementSummary{}, err
	}

	var summary incomeSettlementSummary
	accumulateSettlementSummary(&summary, firstResp.GetRecords())
	totalPages := int((firstResp.GetTotal() + int64(incomePageSize) - 1) / int64(incomePageSize))
	if totalPages <= 1 || len(firstResp.GetRecords()) == 0 {
		return summary, nil
	}
	if totalPages > maxIncomeSettlementPages {
		summary.capped = true
		totalPages = maxIncomeSettlementPages
	}

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	pageCh := make(chan int32)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	var mu sync.Mutex

	worker := func() {
		defer wg.Done()
		for page := range pageCh {
			resp, err := client.ListSettlements(ctx, &payproto.ListSettlementsRequest{
				DriverId: driverID,
				Status:   settlementStatusSettled,
				StartAt:  startAt,
				EndAt:    endAt,
				Page:     page,
				PageSize: incomePageSize,
			})
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}
			mu.Lock()
			accumulateSettlementSummary(&summary, resp.GetRecords())
			mu.Unlock()
		}
	}

	workerCount := incomePageFetchWorkers
	if remaining := totalPages - 1; remaining < workerCount {
		workerCount = remaining
	}
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	go func() {
		defer close(pageCh)
		for page := int32(2); int(page) <= totalPages; page++ {
			select {
			case <-ctx.Done():
				return
			case pageCh <- page:
			}
		}
	}()

	wg.Wait()
	select {
	case err := <-errCh:
		return incomeSettlementSummary{}, err
	default:
	}
	return summary, nil
}

func accumulateSettlementSummary(summary *incomeSettlementSummary, bills []*payproto.SettlementBill) {
	for _, bill := range bills {
		summary.count++
		summary.incomeCents += bill.GetDriverIncomeCents()
	}
}

func settlementDisplayTime(bill *payproto.SettlementBill) int64 {
	if bill == nil {
		return 0
	}
	if bill.GetSettledAt() > 0 {
		return bill.GetSettledAt()
	}
	return bill.GetCreatedAt()
}

func (l *IncomeLogic) payClient() (svc.PayClient, error) {
	if l.svcCtx == nil || l.svcCtx.PayClient == nil {
		return nil, ErrPayClientNotConfigured
	}
	return l.svcCtx.PayClient, nil
}
