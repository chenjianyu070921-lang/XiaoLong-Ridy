package driver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// Agent 网约车司机自动评价 Agent：
// 模型判定优先，输出经严格校验（星级边界 / 标签白名单 / 理由 100 字上限），
// 校验不过或模型不可用时降级到内置规则引擎（仅依据已提供的事实，不编造行为）。
type Agent struct {
	model     ModelClient
	knowledge KnowledgeRetriever
}

// NewAgent 构造自动评价 Agent；model 为 nil 时直接走规则引擎，knowledge 为 nil 时视为知识库为空。
func NewAgent(model ModelClient, knowledge KnowledgeRetriever) *Agent {
	return &Agent{model: model, knowledge: knowledge}
}

// Rate 对已完成订单打分，返回固定结构结果。
func (a *Agent) Rate(ctx context.Context, order OrderInput) (*RatingResult, error) {
	if strings.TrimSpace(order.OrderID) == "" {
		return nil, errors.New("order_id required")
	}

	user := buildUserPrompt(retrieveKnowledge(ctx, a.knowledge, order), order)
	if a.model != nil {
		raw, err := a.model.Generate(ctx, systemPrompt, user)
		if err != nil {
			// 模型失败不中断打分主链路，降级到规则引擎。
			raw = ""
		}
		if r := validateRating(raw); r != nil {
			return r, nil
		}
	}
	return fallbackRating(order), nil
}

// RateByDriverPhone 按司机手机号查询乘客评价并自动打分：
// 从乘客评价储存表（passenger_review）读取该司机最近的乘客文字评价，
// 拼接为乘客评价输入（评分标准约束 3：优先乘客文本评价），再走 Rate 主链路
// （模型判定 + 输出校验 + 规则降级）。
// 当前轨迹/行为记录暂无独立数据源，留空，由评分规则按"只依据乘客评价"执行。
func (a *Agent) RateByDriverPhone(ctx context.Context, store ReviewStore, driverPhone string, maxReviews int) (*RatingResult, error) {
	if store == nil {
		return nil, errors.New("review store required")
	}
	driverPhone = strings.TrimSpace(driverPhone)
	if driverPhone == "" {
		return nil, errors.New("driver phone required")
	}
	if maxReviews <= 0 {
		maxReviews = 20
	}
	reviews, err := store.ListByDriverPhone(ctx, driverPhone, maxReviews)
	if err != nil {
		return nil, err
	}
	return a.Rate(ctx, OrderInput{
		OrderID:          "driver-phone:" + driverPhone,
		PassengerComment: joinPassengerComments(reviews),
	})
}

// joinPassengerComments 将多条乘客评价拼接为一段评价文本；
// 单条截断 80 字、总长截断 600 字，避免超长输入拖垮提示词。
func joinPassengerComments(reviews []PassengerReview) string {
	parts := make([]string, 0, len(reviews))
	for _, r := range reviews {
		comment := truncateRunes(r.Comment, 80)
		if comment == "" {
			continue
		}
		parts = append(parts, comment)
	}
	return truncateRunes(strings.Join(parts, "；"), 600)
}

// validateRating 解析并校验模型输出：
// 必须是合法 JSON、星级在 1-5、标签必须在白名单集合内（非法标签丢弃），
// 任一硬条件不满足返回 nil 触发规则降级；理由超过 100 字截断。
func validateRating(raw string) *RatingResult {
	var r RatingResult
	if err := json.Unmarshal([]byte(trimJSON(raw)), &r); err != nil {
		return nil
	}
	if r.Star < 1 || r.Star > 5 {
		return nil
	}
	tags := filterTags(r.Tags)
	return &RatingResult{
		Star:   r.Star,
		Tags:   tags,
		Reason: truncateReason(r.Reason),
		Source: SourceLLM,
	}
}

// filterTags 依据白名单过滤标签并去重；无标签时返回空数组而非 nil（强制约束 2）。
func filterTags(raw []string) []string {
	allowed := make(map[string]struct{}, len(TagWhitelist))
	for _, t := range TagWhitelist {
		allowed[t] = struct{}{}
	}
	tags := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if _, ok := allowed[t]; !ok {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		tags = append(tags, t)
	}
	return tags
}

// truncateReason 将判定理由截断到 100 字（rune）以内。
func truncateReason(s string) string {
	return truncateRunes(s, 100)
}

// truncateRunes 将字符串截断到 max 个 rune 以内。
func truncateRunes(s string, max int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) > max {
		runes = runes[:max]
	}
	return string(runes)
}

// trimJSON 去除模型输出中可能的 markdown 代码块包裹与前后噪声。
func trimJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if i := strings.LastIndex(s, "}"); i >= 0 {
		s = s[:i+1]
	}
	return s
}
