package driver

import "strings"

// severityRank 规则引擎的严重档位，与星级判定标准一一对应。
type severityRank int

const (
	rankNone     severityRank = iota // 5星：无任何负面行为
	rankSlight                       // 4星：轻微小问题
	rankObvious                      // 3星：明显瑕疵
	rankSevere                       // 2星：明显违规
	rankCritical                     // 1星：严重问题
)

func (r severityRank) star() int {
	switch r {
	case rankCritical:
		return 1
	case rankSevere:
		return 2
	case rankObvious:
		return 3
	case rankSlight:
		return 4
	default:
		return 5
	}
}

// fallbackRating 内置规则引擎：仅依据调用方提供的乘客评价文本与结构化行为记录
// 做确定性打分，不调用模型、不编造任何未提供的行为（强制约束 6）。
// 判定顺序遵循"优先乘客文本评价，再结合订单行为记录"（强制约束 3）。
func fallbackRating(order OrderInput) *RatingResult {
	rank, tags := scoreFromComment(order.PassengerComment)
	brank, btags := scoreFromBehaviors(order.Behaviors)
	if brank > rank {
		rank = brank
	}
	tags = filterTags(append(tags, btags...))

	// 无严重/明显违规时，补充有据可查的正向标签。
	if rank <= rankObvious {
		tags = append(tags, positiveTags(order)...)
		tags = filterTags(tags)
	}

	return &RatingResult{
		Star:   rank.star(),
		Tags:   tags,
		Reason: buildRuleReason(order, rank, tags),
		Source: SourceRuleFallback,
	}
}

// scoreFromBehaviors 汇总行为记录的严重档位与标签；事件严重程度可覆盖默认档位。
func scoreFromBehaviors(events []BehaviorEvent) (severityRank, []string) {
	rank := rankNone
	tags := make([]string, 0, len(events))
	for _, ev := range events {
		r, hasTag, tag := mapBehavior(ev)
		if r > rank {
			rank = r
		}
		if hasTag {
			tags = append(tags, tag)
		}
	}
	return rank, tags
}

// mapBehavior 单条行为记录 →（严重档位, 是否有对应白名单标签, 标签）。
func mapBehavior(ev BehaviorEvent) (severityRank, bool, string) {
	// 严重问题无论标注什么档位都按严重问题处理。
	switch ev.Name {
	case BehaviorInduceCancel:
		return rankCritical, true, TagInduceCancel
	case BehaviorVerbalAbuse:
		return rankCritical, true, TagBadAttitude
	case BehaviorDangerousDriving:
		return rankCritical, false, ""
	case BehaviorDetour:
		switch ev.Severity {
		case SeveritySlight:
			return rankObvious, true, TagDetour
		default:
			return rankSevere, true, TagDetour // 绕路属明显违规（2星档）
		}
	case BehaviorLatePickup:
		switch ev.Severity {
		case SeveritySlight:
			return rankSlight, true, TagLatePickup
		case SeveritySevere:
			return rankSevere, true, TagLatePickup
		default:
			return rankObvious, true, TagLatePickup // 接驾慢属明显瑕疵（3星档）
		}
	case BehaviorBadAttitude:
		return rankSevere, true, TagBadAttitude
	case BehaviorDirtyCar:
		switch ev.Severity {
		case SeveritySevere:
			return rankSevere, true, TagDirtyCar
		default:
			return rankObvious, true, TagDirtyCar
		}
	case BehaviorNoAnswer:
		switch ev.Severity {
		case SeveritySevere:
			return rankSevere, true, TagNoAnswer
		default:
			return rankObvious, true, TagNoAnswer
		}
	case BehaviorEarlyArrive:
		return rankObvious, true, TagEarlyArrive
	default:
		// 未登记的行为名一律忽略，防止调用方传入未知事实。
		return rankNone, false, ""
	}
}

// scoreFromComment 从乘客文字评价中提取严重档位与标签。
func scoreFromComment(comment string) (severityRank, []string) {
	rank := rankNone
	tags := make([]string, 0, 4)
	add := func(r severityRank, tag string) {
		if r > rank {
			rank = r
		}
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	if containsAny(comment, "诱导", "劝我取消", "让我取消") && containsAny(comment, "取消") {
		add(rankCritical, TagInduceCancel)
	}
	if containsAny(comment, "辱骂", "骂人", "吵架") {
		add(rankCritical, TagBadAttitude)
	}
	if containsAny(comment, "态度差", "态度恶劣", "态度很凶", "态度很差") {
		add(rankSevere, TagBadAttitude)
	}
	if containsAny(comment, "危险驾驶", "闯红灯", "猛踩", "急刹") {
		add(rankCritical, "")
	}
	if containsAny(comment, "严重绕路") {
		add(rankSevere, TagDetour)
	} else if containsAny(comment, "绕路", "多收") {
		add(rankObvious, TagDetour)
	}
	if containsAny(comment, "严重超时", "等了很久", "迟到", "接驾慢") {
		add(rankObvious, TagLatePickup)
	}
	if containsAny(comment, "不接电话", "联系不上", "打不通") {
		add(rankObvious, TagNoAnswer)
	}
	if containsAny(comment, "提前点到达", "没到就点", "还没到就") {
		add(rankObvious, TagEarlyArrive)
	}
	if containsAny(comment, "脏", "异味", "臭", "烟味") {
		add(rankObvious, TagDirtyCar)
	}
	return rank, tags
}

// positiveTags 输出有据可查的正向标签：
// 仅来自明确的正向行为记录或乘客正面评价；缺事实时保持空数组，不凭空给好评标签。
func positiveTags(order OrderInput) []string {
	tags := make([]string, 0, 4)
	comment := order.PassengerComment
	has := func(names ...string) bool {
		for _, ev := range order.Behaviors {
			for _, n := range names {
				if ev.Name == n {
					return true
				}
			}
		}
		return false
	}
	if has(BehaviorOnTime) || containsAny(comment, "准时") {
		tags = append(tags, TagOnTime)
	}
	if has(BehaviorCleanCar) || containsAny(comment, "干净", "整洁") {
		tags = append(tags, TagCleanCar)
	}
	if containsAny(comment, "态度好", "热情", "礼貌") {
		tags = append(tags, TagGoodAttitude)
	}
	return tags
}

// buildRuleReason 生成简短判定理由，只引用已提供的事实，超长截断到 100 字。
func buildRuleReason(order OrderInput, rank severityRank, tags []string) string {
	var parts []string
	if c := strings.TrimSpace(order.PassengerComment); c != "" {
		parts = append(parts, "乘客评价："+truncateRunes(c, 40))
	}
	if len(order.Behaviors) > 0 {
		names := make([]string, 0, len(order.Behaviors))
		for _, ev := range order.Behaviors {
			names = append(names, behaviorLabel(ev))
		}
		parts = append(parts, "行为记录："+strings.Join(names, "、"))
	}
	var sb strings.Builder
	if len(parts) > 0 {
		sb.WriteString(strings.Join(parts, "；"))
		sb.WriteString("。")
	} else {
		sb.WriteString("订单轨迹与行为记录无负面行为。")
	}
	if len(tags) > 0 {
		sb.WriteString("标签：")
		sb.WriteString(strings.Join(tags, "、"))
		sb.WriteString("。")
	}
	sb.WriteString("依据内置规则判定")
	sb.WriteString(rankLabel(rank))
	sb.WriteString("。")
	return truncateReason(sb.String())
}

func behaviorLabel(ev BehaviorEvent) string {
	switch ev.Name {
	case BehaviorLatePickup:
		return "接驾迟到"
	case BehaviorDetour:
		return "绕路"
	case BehaviorBadAttitude:
		return "态度差"
	case BehaviorDirtyCar:
		return "车内脏乱"
	case BehaviorInduceCancel:
		return "诱导取消"
	case BehaviorNoAnswer:
		return "不接电话"
	case BehaviorEarlyArrive:
		return "提前点到达"
	case BehaviorVerbalAbuse:
		return "辱骂"
	case BehaviorDangerousDriving:
		return "危险驾驶"
	case BehaviorOnTime:
		return "接驾准时"
	case BehaviorCleanCar:
		return "车内干净"
	default:
		return ev.Name
	}
}

func rankLabel(r severityRank) string {
	return itoa(r.star()) + "星"
}

func itoa(n int) string {
	if n < 0 || n > 9 {
		return "?"
	}
	return string(rune('0' + n))
}

func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if k != "" && strings.Contains(s, k) {
			return true
		}
	}
	return false
}
