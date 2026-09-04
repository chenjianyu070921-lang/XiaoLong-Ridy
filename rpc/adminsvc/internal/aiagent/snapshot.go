package aiagent

// demoFacts 返回指定场景的内置演示快照事实。
//
// 快照是版本化只读数据，仅答辩/本地环境且显式开启 demo_mode 时使用；
// 它不写回业务表，也不参与真实生产处置。订单号与风险 ID 固定，保证跨轮引用稳定可复现。
func demoFacts(scene Scene) *facts {
	switch scene {
	case SceneOverview:
		return &facts{Metrics: &MetricsFact{
			OrderCount:      1284,
			CompletedCount:  1045,
			AbnormalCount:   23,
			CancelRate:      "18.6%",
			CompletionRate:  "81.4%",
			TimeoutCount:    14,
			PaymentAbnormal: 9,
			GMV:             "68420.00",
			UserCount:       8921,
			DriverCount:     320,
			DataAsOf:        "2026-09-01 18:00:00",
		}}
	case SceneAbnormalOrder:
		return &facts{Abnormal: []AbnormalOrderFact{
			{OrderNo: "202609010001", AbnormalType: "dispatch", AbnormalReason: "派单等待超时", PaymentStatus: 1, DispatchStatus: 3, Status: 1},
			{OrderNo: "202609010002", AbnormalType: "payment", AbnormalReason: "支付超时未到账", PaymentStatus: 3, DispatchStatus: 2, Status: 4},
			{OrderNo: "202609010003", AbnormalType: "cancel", AbnormalReason: "重复取消", PaymentStatus: 1, DispatchStatus: 5, Status: 6},
		}}
	case SceneRiskReview:
		return &facts{Risks: []RiskHitFact{
			{ID: 9001, TargetType: "user", TargetID: 1024, Scene: "login", RiskLevel: 3, HitReason: "短时间多设备登录", HandleStatus: "pending"},
			{ID: 9002, TargetType: "driver", TargetID: 66, Scene: "order", RiskLevel: 2, HitReason: "高频取消订单", HandleStatus: "pending"},
			{ID: 9003, TargetType: "phone", TargetID: 8801, Scene: "register", RiskLevel: 3, HitReason: "命中黑名单号码", HandleStatus: "pending"},
		}}
	default:
		return &facts{}
	}
}
