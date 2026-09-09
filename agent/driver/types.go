// Package driver 实现网约车司机自动评价 Agent：
// 严格遵循花小猪打车乘客评价司机的评分标准，对已完成订单自动打分，
// 输出固定 JSON（star/tags/reason），标签只允许来自白名单集合。
package driver

import "context"

// 正向标签白名单。
const (
	TagOnTime          = "接驾准时"
	TagReasonableRoute = "路线合理"
	TagGoodAttitude    = "服务态度好"
	TagCleanCar        = "车内干净"
)

// 负向标签白名单。
const (
	TagLatePickup   = "接驾迟到"
	TagDetour       = "司机绕路"
	TagBadAttitude  = "司机态度差"
	TagDirtyCar     = "车内脏乱"
	TagInduceCancel = "诱导取消"
	TagNoAnswer     = "不接电话"
	TagEarlyArrive  = "提前点到达"
)

// TagWhitelist 是允许输出的全部标签（正向 + 负向），超出集合的标签一律丢弃。
var TagWhitelist = []string{
	TagOnTime, TagReasonableRoute, TagGoodAttitude, TagCleanCar,
	TagLatePickup, TagDetour, TagBadAttitude, TagDirtyCar,
	TagInduceCancel, TagNoAnswer, TagEarlyArrive,
}

// 行为事件名，来源于订单轨迹日志与接驾/行驶行为记录。
const (
	BehaviorLatePickup       = "late_pickup"       // 接驾迟到/严重超时
	BehaviorDetour           = "detour"            // 绕路/偏航
	BehaviorBadAttitude      = "bad_attitude"      // 态度差/争吵/言语冒犯
	BehaviorDirtyCar         = "dirty_car"         // 车内脏乱/异味
	BehaviorInduceCancel     = "induce_cancel"     // 诱导乘客取消订单
	BehaviorNoAnswer         = "no_answer_phone"   // 多次不接乘客电话
	BehaviorEarlyArrive      = "early_arrive"      // 提前点到达
	BehaviorVerbalAbuse      = "verbal_abuse"      // 辱骂
	BehaviorDangerousDriving = "dangerous_driving" // 危险驾驶
	BehaviorOnTime           = "on_time_pickup"    // 接驾准时（正向）
	BehaviorCleanCar         = "clean_car"         // 车内干净（正向）
)

// Severity 行为严重程度；规则引擎据此映射星级档位。
type Severity string

const (
	SeveritySlight Severity = "slight" // 轻微小问题
	SeverityNormal Severity = "normal" // 默认档位
	SeveritySevere Severity = "severe" // 明显违规/严重
)

// BehaviorEvent 一条已确认的司机行为记录，禁止传入未发生的行为。
type BehaviorEvent struct {
	Name     string   `json:"name"`
	Severity Severity `json:"severity,omitempty"`
	Detail   string   `json:"detail,omitempty"`
}

// TrackPoint 订单轨迹点（调用方应先采样，避免超长输入）。
type TrackPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

// OrderInput 对应 order_info：订单轨迹、接驾时长、行驶路线、乘客评价文本。
type OrderInput struct {
	OrderID           string          `json:"order_id"`
	PickupDurationSec int64           `json:"pickup_duration_sec,omitempty"`
	Track             []TrackPoint    `json:"track,omitempty"`
	PassengerComment  string          `json:"passenger_comment,omitempty"`
	Behaviors         []BehaviorEvent `json:"behaviors,omitempty"`
}

// 结论来源标记。
const (
	SourceLLM          = "llm"           // 模型判定且通过校验
	SourceRuleFallback = "rule_fallback" // 模型不可用/输出非法时，内置规则引擎判定
)

// RatingResult 打分结果，star/tags/reason 字段与强制输出格式一致；
// Source 为平台侧审计字段，不进入对外评价 JSON。
type RatingResult struct {
	Star   int      `json:"star"`
	Tags   []string `json:"tags"`
	Reason string   `json:"reason"`
	Source string   `json:"source"`
}

// ModelClient 可替换的 LLM 边界，与 adminsvc aiagent 的客户端同形，便于复用适配器。
type ModelClient interface {
	Generate(ctx context.Context, system, user string) (string, error)
}

// KnowledgeRetriever 按订单检索知识库参考片段；
// 返回空串表示知识库为空，Agent 只使用内置评分规则。
type KnowledgeRetriever interface {
	Retrieve(ctx context.Context, order OrderInput) (string, error)
}
