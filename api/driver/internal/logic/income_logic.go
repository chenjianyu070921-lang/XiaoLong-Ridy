package logic

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	payproto "XiaoLong-Ridy/rpc/paysvc/proto"
)

// errDriverClientNotConfigured 表示 ServiceContext 未注入司机 RPC 客户端。
var errDriverClientNotConfigured = errors.New("driver client not configured")

const (
	incomePageSize           int32 = 100
	incomePageFetchWorkers         = 8
	maxIncomeSettlementPages       = 20
	incomeSource                   = "paysvc.settlement"
	settlementStatusSettled  int32 = 2
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
	// P1-1 修复：可提现额必须扣减「已申请 + 已打款」的提现金额，
	// 否则司机看到的可提余额虚高，可超量/重复提现造成资损。
	withdrawnCents, err := l.sumWithdrawnCents(driverID)
	if err != nil {
		return nil, err
	}
	withdrawable := summary.incomeCents - withdrawnCents
	if withdrawable < 0 {
		withdrawable = 0
	}
	return &types.GetIncomeSummaryResponse{
		DriverID:          driverID,
		CompletedOrders:   summary.count,
		TotalIncomeCents:  summary.incomeCents,
		WithdrawableCents: withdrawable,
		Source:            source,
	}, nil
}

// sumWithdrawnCents 累加司机「已申请(pending=1) + 已打款(paid=2)」提现金额，换算为分。
// 这两类都占用可提余额：已打款已实际到账，已申请尚未处理但不可重复提取。
func (l *IncomeLogic) sumWithdrawnCents(driverID int64) (int64, error) {
	client := l.svcCtx.DriverClient
	if client == nil {
		return 0, errDriverClientNotConfigured
	}
	const pageSize int32 = 100
	var (
		page  int32 = 1
		total int64
	)
	for {
		resp, err := client.ListWithdraws(l.ctx, &driversproto.ListWithdrawsRequest{
			DriverId: driverID,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return 0, err
		}
		for _, w := range resp.GetRecords() {
			if w.GetStatus() == int32(1) || w.GetStatus() == int32(2) {
				total += int64(math.Round(w.GetAmount() * 100))
			}
		}
		if int64(len(resp.GetRecords())) < int64(pageSize) {
			break
		}
		page++
		if page > 1000 { // 安全上限，避免异常分页导致死循环
			break
		}
	}
	return total, nil
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
