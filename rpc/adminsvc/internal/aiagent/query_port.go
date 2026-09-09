package aiagent

import "context"

// MetricsFact 是脱敏后的运营指标事实。
// 字段不包含任何个人敏感信息（手机号、身份证、精确地址、支付流水号等）。
type MetricsFact struct {
	OrderCount       int64
	CompletedCount   int64
	AbnormalCount    int64
	CancelRate       string
	CompletionRate   string
	TimeoutCount     int64
	PaymentAbnormal  int64
	GMV              string
	UserCount        int64
	DriverCount      int64
	DataAsOf         string
}

// AbnormalOrderFact 是脱敏后的异常订单事实。
type AbnormalOrderFact struct {
	OrderNo        string
	AbnormalType   string // cancel / payment / dispatch
	AbnormalReason string
	PaymentStatus  int32
	DispatchStatus int32
	Status         int32
}

// RiskHitFact 是脱敏后的风控命中事实。
type RiskHitFact struct {
	ID           int64
	TargetType   string
	TargetID     int64
	Scene        string
	RiskLevel    int32 // 1 低 / 2 中 / 3 高
	HitReason    string
	HandleStatus string // pending / review_pass / blacklisted / work_order
}

// QueryPort 是 aiagent 读取结构化事实的只读端口。
//
// 实现方（adminsvc）只能通过既有下游 RPC 查询，禁止在实现中直读数据库，
// 以避免重蹈历史数据边界穿透问题。
type QueryPort interface {
	// QueryMetrics 返回指定时间范围内的运营指标。
	QueryMetrics(ctx context.Context, startTime, endTime string) (*MetricsFact, error)
	// ListAbnormalOrders 返回指定时间范围内最需要优先处理的异常订单（已按优先级截断）。
	ListAbnormalOrders(ctx context.Context, startTime, endTime string, limit int) ([]AbnormalOrderFact, error)
	// GetRiskSummary 返回待复核的高风险命中记录摘要（已按风险等级排序并截断）。
	GetRiskSummary(ctx context.Context, limit int) ([]RiskHitFact, error)
}
