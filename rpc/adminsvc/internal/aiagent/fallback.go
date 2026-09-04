package aiagent

import (
	"fmt"
	"strconv"
)

// buildTemplateAnswer 基于已核验事实生成本地模板报告，不依赖外部模型。
// 这是四级降级中"真实业务查询 + 本地模板报告"与"演示快照 + 本地模板报告"的实现。
func buildTemplateAnswer(scene Scene, f *facts) *Answer {
	switch scene {
	case SceneOverview:
		return templateOverview(f)
	case SceneAbnormalOrder:
		return templateAbnormal(f)
	case SceneRiskReview:
		return templateRisk(f)
	default:
		return &Answer{Summary: "暂不支持该场景。", SourceMode: SourceTemplateFallback}
	}
}

func templateOverview(f *facts) *Answer {
	m := f.Metrics
	if m == nil {
		m = &MetricsFact{}
	}
	answer := &Answer{
		Summary:    overviewSummary(m),
		Evidence:   overviewEvidence(m),
		Priorities: []Priority{},
		Actions: []Action{
			{Type: "navigate", Label: "查看数据统计", Route: "/statistics"},
			{Type: "navigate", Label: "查看活动配置", Route: "/promotion-activities"},
		},
		SourceMode: SourceTemplateFallback,
		Citations:  []string{"运营总览统计接口"},
	}
	return answer
}

func overviewSummary(m *MetricsFact) string {
	parts := make([]string, 0, 3)
	if m.AbnormalCount > 0 {
		parts = append(parts, fmt.Sprintf("存在 %d 笔异常订单需关注", m.AbnormalCount))
	}
	if m.TimeoutCount > 0 {
		parts = append(parts, fmt.Sprintf("%d 笔派单超时", m.TimeoutCount))
	}
	if m.CancelRate != "" {
		parts = append(parts, "取消率 "+m.CancelRate)
	}
	if len(parts) == 0 {
		return "当前运营平稳，暂无显著异常。"
	}
	return "主要风险：" + joinCN(parts) + "，建议优先核查异常订单。"
}

func overviewEvidence(m *MetricsFact) []Evidence {
	return []Evidence{
		{Label: "订单总量", Value: strconv.FormatInt(m.OrderCount, 10)},
		{Label: "完成率", Value: m.CompletionRate},
		{Label: "取消率", Value: m.CancelRate},
		{Label: "异常订单", Value: strconv.FormatInt(m.AbnormalCount, 10)},
		{Label: "GMV", Value: m.GMV},
		{Label: "用户总量", Value: strconv.FormatInt(m.UserCount, 10)},
		{Label: "司机总量", Value: strconv.FormatInt(m.DriverCount, 10)},
	}
}

func templateAbnormal(f *facts) *Answer {
	orders := f.Abnormal
	if len(orders) == 0 {
		return &Answer{
			Summary:    "当前时间范围内未发现异常订单。",
			Evidence:   []Evidence{{Label: "异常订单", Value: "0"}},
			Priorities: []Priority{},
			Actions: []Action{
				{Type: "navigate", Label: "查看异常订单", Route: "/orders/abnormal"},
			},
			SourceMode: SourceTemplateFallback,
			Citations:  []string{"异常订单查询接口"},
		}
	}

	counts := map[string]int{}
	for _, o := range orders {
		counts[o.AbnormalType]++
	}

	answer := &Answer{
		Summary: fmt.Sprintf("共 %d 笔异常订单，其中派单异常 %d 笔、支付异常 %d 笔、取消异常 %d 笔，建议按优先级复核。",
			len(orders), counts["dispatch"], counts["payment"], counts["cancel"]),
		Evidence: []Evidence{
			{Label: "异常订单总数", Value: strconv.Itoa(len(orders))},
			{Label: "派单异常", Value: strconv.Itoa(counts["dispatch"])},
			{Label: "支付异常", Value: strconv.Itoa(counts["payment"])},
			{Label: "取消异常", Value: strconv.Itoa(counts["cancel"])},
		},
		Priorities: abnormalPriorities(orders),
		Actions: []Action{
			{Type: "navigate", Label: "查看异常订单", Route: "/orders/abnormal"},
			{Type: "navigate", Label: "查看工单", Route: "/work-orders"},
		},
		SourceMode: SourceTemplateFallback,
		Citations:  []string{"异常订单查询接口"},
	}
	return answer
}

func templateRisk(f *facts) *Answer {
	risks := f.Risks
	pending := make([]RiskHitFact, 0, len(risks))
	for _, r := range risks {
		if r.HandleStatus == "pending" {
			pending = append(pending, r)
		}
	}
	if len(risks) == 0 {
		return &Answer{
			Summary:    "当前无待复核的高风险命中记录。",
			Evidence:   []Evidence{{Label: "高风险命中", Value: "0"}},
			Priorities: []Priority{},
			Actions: []Action{
				{Type: "navigate", Label: "查看风控命中", Route: "/risk-hits"},
			},
			SourceMode: SourceTemplateFallback,
			Citations:  []string{"风控命中查询接口"},
		}
	}

	answer := &Answer{
		Summary: fmt.Sprintf("共 %d 条风控命中，其中 %d 条待复核，建议优先处置高风险对象。", len(risks), len(pending)),
		Evidence: []Evidence{
			{Label: "风控命中总数", Value: strconv.Itoa(len(risks))},
			{Label: "待复核", Value: strconv.Itoa(len(pending))},
			{Label: "高风险", Value: strconv.Itoa(countHighRisk(risks))},
		},
		Priorities: riskPriorities(risks),
		Actions: []Action{
			{Type: "navigate", Label: "查看风控命中", Route: "/risk-hits"},
			{Type: "navigate", Label: "查看黑名单", Route: "/blacklist"},
		},
		SourceMode: SourceTemplateFallback,
		Citations:  []string{"风控命中与黑名单查询接口"},
	}
	return answer
}

// abnormalPriorities 将异常订单截断为最多 3 条优先对象，并映射等级与路由。
func abnormalPriorities(orders []AbnormalOrderFact) []Priority {
	limit := len(orders)
	if limit > 3 {
		limit = 3
	}
	out := make([]Priority, 0, limit)
	for i := 0; i < limit; i++ {
		o := orders[i]
		out = append(out, Priority{
			Type:    "order",
			ID:      o.OrderNo,
			Level:   abnormalLevel(o.AbnormalType),
			Reasons: []string{abnormalTypeText(o.AbnormalType)},
			Route:   "/orders/abnormal",
		})
	}
	return out
}

// riskPriorities 将风控命中截断为最多 3 条优先对象。
func riskPriorities(risks []RiskHitFact) []Priority {
	limit := len(risks)
	if limit > 3 {
		limit = 3
	}
	out := make([]Priority, 0, limit)
	for i := 0; i < limit; i++ {
		r := risks[i]
		out = append(out, Priority{
			Type:    "risk",
			ID:      riskIDString(r.ID),
			Level:   riskLevelText(r.RiskLevel),
			Reasons: []string{r.HitReason},
			Route:   "/risk-hits",
		})
	}
	return out
}

func abnormalLevel(t string) string {
	switch t {
	case "dispatch", "payment":
		return "high"
	default:
		return "medium"
	}
}

func abnormalTypeText(t string) string {
	switch t {
	case "cancel":
		return "取消异常"
	case "payment":
		return "支付异常"
	case "dispatch":
		return "派单异常"
	default:
		return "异常"
	}
}

func riskLevelText(level int32) string {
	switch level {
	case 3:
		return "high"
	case 2:
		return "medium"
	default:
		return "low"
	}
}

func countHighRisk(risks []RiskHitFact) int {
	n := 0
	for _, r := range risks {
		if r.RiskLevel >= 3 {
			n++
		}
	}
	return n
}

// joinCN 用顿号连接中文短语。
func joinCN(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += "、"
		}
		out += s
	}
	return out
}
