// Package aiagent 是管理后台 AI 运营助手的核心编排模块。
//
// 本模块不依赖 RPC 传输层：通过 QueryPort 注入既有只读查询能力，
// 通过 ModelClient 注入外部 LLM，通过 SessionStore 注入会话存储，
// 从而可以独立单元测试，并在模型不可用时降级到本地模板报告。
//
// 安全边界（与设计文档一致）：
//   - 模型不能生成 SQL、不能访问数据库、不能调用写接口。
//   - 事实只来自 QueryPort（服务端白名单查询）或内置演示快照。
//   - 回答中的证据 ID 与跳转对象必须来自本次工具结果，服务端负责校验。
package aiagent

// Scene 表示受限问答场景。
type Scene string

const (
	// SceneOverview 运营数据问答。
	SceneOverview Scene = "overview"
	// SceneAbnormalOrder 异常订单分析。
	SceneAbnormalOrder Scene = "abnormal_order"
	// SceneRiskReview 风控审核助手。
	SceneRiskReview Scene = "risk_review"
)

// Valid 判断场景是否合法。
func (s Scene) Valid() bool {
	switch s {
	case SceneOverview, SceneAbnormalOrder, SceneRiskReview:
		return true
	default:
		return false
	}
}

// SourceMode 表示回答的数据来源模式，前端据此显示“实时数据/演示快照/本地规则报告”。
type SourceMode string

const (
	// SourceRealtime 真实业务查询 + 外部 LLM。
	SourceRealtime SourceMode = "realtime"
	// SourceDemoSnapshot 演示快照数据（答辩环境）。
	SourceDemoSnapshot SourceMode = "demo_snapshot"
	// SourceTemplateFallback 真实业务查询 + 本地模板报告。
	SourceTemplateFallback SourceMode = "template_fallback"
)

// Evidence 是回答中的一条数据证据。
type Evidence struct {
	Label      string `json:"label"`
	Value      string `json:"value"`
	Comparison string `json:"comparison,omitempty"`
}

// Priority 是需要优先处理的订单或风控记录。
type Priority struct {
	Type    string   `json:"type"`            // order / risk
	ID      string   `json:"id"`              // 订单号或风控记录号
	Level   string   `json:"level"`           // high / medium / low
	Reasons []string `json:"reasons"`         // 触发因素
	Route   string   `json:"route"`           // 详情页跳转路径
}

// Action 是建议动作，仅支持页面跳转或草稿，绝不执行写操作。
type Action struct {
	Type  string `json:"type"` // navigate
	Label string `json:"label"`
	Route string `json:"route"`
}

// Answer 是服务端校验后的标准化回答。
type Answer struct {
	Summary        string     `json:"summary"`
	Evidence       []Evidence `json:"evidence"`
	Priorities     []Priority `json:"priorities"`
	Actions        []Action   `json:"actions"`
	SourceMode     SourceMode `json:"source_mode"`
	Citations      []string   `json:"citations"`
	ConversationID string     `json:"conversation_id"`
	TraceID        string     `json:"trace_id"`
}

// Suggestion 是场景快捷问题。
type Suggestion struct {
	Scene       Scene  `json:"scene"`
	QuickPrompt string `json:"quick_prompt"`
}

// Config 是 aiagent 模块的运行时配置，由 adminsvc 配置层注入。
type Config struct {
	// DemoEnabled 是否允许演示快照；生产环境必须为 false。
	DemoEnabled bool
	// Model 是外部 LLM 配置；为空时走本地模板降级。
	Model ModelConfig
	// Conversation 会话上下文管理参数。
	Conversation ConversationConfig
}

// ModelConfig 描述外部 LLM 连接参数。密钥不在此结构内，由 ModelClient 实现方从环境变量读取。
type ModelConfig struct {
	// Endpoint 模型供应商地址；为空表示未配置模型。
	Endpoint string
	// Name 模型名。
	Name string
	// TimeoutSeconds 单次调用超时。
	TimeoutSeconds int
}

// ConversationConfig 描述会话上下文管理参数。
type ConversationConfig struct {
	// TTLSeconds 会话无活动过期时间，默认 24 小时。
	TTLSeconds int
	// MaxRounds 保留的最近轮次，默认 6。
	MaxRounds int
	// KeyPrefix Redis key 前缀。
	KeyPrefix string
}
