package aiagent

import "strconv"

// facts 是编排层在单次回答中已核验的结构化事实集合。
// 它既是本地模板报告的数据源，也是校验模型输出（证据 ID 必须来自本次结果）的依据。
type facts struct {
	Metrics  *MetricsFact         `json:"metrics,omitempty"`
	Abnormal []AbnormalOrderFact  `json:"abnormal_orders,omitempty"`
	Risks    []RiskHitFact        `json:"risk_hits,omitempty"`
}

// priorityIDs 返回本次事实中所有可作为优先对象的标识集合，用于校验模型输出。
func (f *facts) priorityIDs() map[string]bool {
	ids := make(map[string]bool)
	for _, o := range f.Abnormal {
		if o.OrderNo != "" {
			ids[o.OrderNo] = true
		}
	}
	for _, r := range f.Risks {
		// 风控命中使用记录 ID 作为标识。
		ids[riskIDString(r.ID)] = true
	}
	return ids
}

// riskIDString 将风控命中记录 ID 转为字符串标识。
func riskIDString(id int64) string {
	if id <= 0 {
		return ""
	}
	return "risk-" + strconv.FormatInt(id, 10)
}
