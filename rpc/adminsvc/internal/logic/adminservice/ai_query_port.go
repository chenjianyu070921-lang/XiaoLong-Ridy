package adminservicelogic

import (
	"context"
	"sort"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/aiagent"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
)

// adminQueryPort 实现 aiagent.QueryPort，复用 adminsvc 既有只读查询逻辑。
// 它不新增任何数据库直读，只调用已存在的查询 logic，避免 AI 助手绕过既有查询边界。
type adminQueryPort struct {
	svcCtx *svc.ServiceContext
}

// QueryMetrics 聚合运营总览与订单统计为脱敏指标事实。
func (p adminQueryPort) QueryMetrics(ctx context.Context, startTime, endTime string) (*aiagent.MetricsFact, error) {
	req := &adminsvc.StatisticsRequest{StartTime: startTime, EndTime: endTime}
	overview, err := NewGetStatisticsOverviewLogic(ctx, p.svcCtx).GetStatisticsOverview(req)
	if err != nil {
		return nil, err
	}
	orderStats, err := NewGetOrderStatisticsLogic(ctx, p.svcCtx).GetOrderStatistics(req)
	if err != nil {
		return nil, err
	}
	return &aiagent.MetricsFact{
		OrderCount:      overview.GetOrderCount(),
		CompletedCount:  overview.GetCompletedOrderCount(),
		AbnormalCount:   overview.GetAbnormalOrderCount(),
		CancelRate:      orderStats.GetCancelRate(),
		CompletionRate:  orderStats.GetCompletionRate(),
		TimeoutCount:    orderStats.GetTimeoutOrderCount(),
		PaymentAbnormal: orderStats.GetPaymentAbnormalCount(),
		GMV:             overview.GetGmv(),
		UserCount:       overview.GetUserCount(),
		DriverCount:     overview.GetDriverCount(),
		DataAsOf:        overview.GetDataAsOf(),
	}, nil
}

// ListAbnormalOrders 返回指定时间范围内的异常订单事实，最多 limit 条。
func (p adminQueryPort) ListAbnormalOrders(ctx context.Context, startTime, endTime string, limit int) ([]aiagent.AbnormalOrderFact, error) {
	if limit <= 0 {
		limit = 3
	}
	resp, err := NewListAbnormalOrdersLogic(ctx, p.svcCtx).ListAbnormalOrders(&adminsvc.AbnormalOrderListRequest{
		Page:      1,
		PageSize:  int32(limit),
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return nil, err
	}
	out := make([]aiagent.AbnormalOrderFact, 0, len(resp.GetList()))
	for _, o := range resp.GetList() {
		out = append(out, aiagent.AbnormalOrderFact{
			OrderNo:        o.GetOrderNo(),
			AbnormalType:   o.GetAbnormalType(),
			AbnormalReason: o.GetAbnormalReason(),
			PaymentStatus:  o.GetPaymentStatus(),
			DispatchStatus: o.GetDispatchStatus(),
			Status:         o.GetStatus(),
		})
	}
	return out, nil
}

// GetRiskSummary 返回待复核的高风险命中事实，按风险等级降序并截断到 limit。
func (p adminQueryPort) GetRiskSummary(ctx context.Context, limit int) ([]aiagent.RiskHitFact, error) {
	if limit <= 0 {
		limit = 3
	}
	resp, err := NewListRiskHitRecordsLogic(ctx, p.svcCtx).ListRiskHitRecords(&adminsvc.RiskHitRecordListRequest{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		return nil, err
	}
	pending := make([]aiagent.RiskHitFact, 0)
	for _, r := range resp.GetList() {
		if r.GetHandleStatus() != "pending" {
			continue
		}
		pending = append(pending, aiagent.RiskHitFact{
			ID:           r.GetId(),
			TargetType:   r.GetTargetType(),
			TargetID:     r.GetTargetId(),
			Scene:        r.GetScene(),
			RiskLevel:    r.GetRiskLevel(),
			HitReason:    r.GetHitReason(),
			HandleStatus: r.GetHandleStatus(),
		})
	}
	sort.SliceStable(pending, func(i, j int) bool {
		return pending[i].RiskLevel > pending[j].RiskLevel
	})
	if len(pending) > limit {
		pending = pending[:limit]
	}
	return pending, nil
}
