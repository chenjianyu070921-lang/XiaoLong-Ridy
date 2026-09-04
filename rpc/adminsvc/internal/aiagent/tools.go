package aiagent

// 工具白名单：模型可请求的工具名，全部由服务端白名单控制。
// 服务端依据场景允许的工具集合拒绝未授权工具，并记录安全事件摘要。
const (
	toolQueryMetrics    = "query_operating_metrics"
	toolListAbnormal    = "list_abnormal_orders"
	toolGetRiskSummary  = "get_risk_hit_summary"
	toolLoadDemoSnapshot = "load_demo_snapshot"
)

// toolIsAllowed 判断某工具是否在该场景白名单内。
func toolIsAllowed(scene Scene, tool string) bool {
	for _, t := range allowedToolsFor(scene) {
		if t == tool {
			return true
		}
	}
	return false
}
