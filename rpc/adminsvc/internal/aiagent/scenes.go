package aiagent

import "errors"

// ErrOutOfScope 表示问题超出三个受限场景或后台业务范围。
var ErrOutOfScope = errors.New("question out of supported scope")

// sceneConfig 描述单个场景的契约：快捷问题、允许工具、跳转页面。
type sceneConfig struct {
	scene       Scene
	quickPrompt string
	// tools 是该场景允许的工具白名单。
	tools []string
	// routes 是该场景可跳转的既有页面。
	routes []string
}

// sceneConfigs 是三个受限场景的固定契约。
// 自由问题仅限这三个场景；无法分类的问题返回范围提示，不作为开放聊天处理。
var sceneConfigs = map[Scene]sceneConfig{
	SceneOverview: {
		scene:       SceneOverview,
		quickPrompt: "分析今日运营风险与增长机会",
		tools:       []string{toolQueryMetrics},
		routes:      []string{"/statistics", "/promotion-activities"},
	},
	SceneAbnormalOrder: {
		scene:       SceneAbnormalOrder,
		quickPrompt: "识别需要优先处理的异常订单",
		tools:       []string{toolListAbnormal},
		routes:      []string{"/orders/abnormal", "/work-orders"},
	},
	SceneRiskReview: {
		scene:       SceneRiskReview,
		quickPrompt: "总结待复核高风险命中并给出建议",
		tools:       []string{toolGetRiskSummary},
		routes:      []string{"/risk-hits", "/blacklist", "/driver-certifications"},
	},
}

// SceneOf 解析并校验场景字符串，非法返回空字符串。
func SceneOf(s string) Scene {
	scene := Scene(s)
	if !scene.Valid() {
		return ""
	}
	return scene
}

// Suggestions 返回三个场景的快捷问题，供前端渲染快捷入口。
func Suggestions() []Suggestion {
	// 按固定顺序返回，保证前端展示稳定。
	order := []Scene{SceneOverview, SceneAbnormalOrder, SceneRiskReview}
	out := make([]Suggestion, 0, len(order))
	for _, s := range order {
		out = append(out, Suggestion{Scene: s, QuickPrompt: sceneConfigs[s].quickPrompt})
	}
	return out
}

// allowedToolsFor 返回场景允许的工具集合；场景非法返回 nil。
func allowedToolsFor(scene Scene) []string {
	cfg, ok := sceneConfigs[scene]
	if !ok {
		return nil
	}
	return cfg.tools
}
