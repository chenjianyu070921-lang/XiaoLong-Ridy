// Package rule 提供支付/结算/退款相关的纯计算与校验逻辑。
package rule

import "math"

// CalcSettlement 计算司机结算：平台抽成 + 司机实收。
//
// totalCents     订单实际总金额（分）
// commissionRate 平台抽成比例（%），如 20 表示 20%
//
// 返回 platformCommissionCents（平台抽成，分）、driverIncomeCents（司机实收，分）。
func CalcSettlement(totalCents int64, commissionRate float64) (int64, int64) {
	if totalCents <= 0 {
		return 0, 0
	}
	if commissionRate < 0 {
		commissionRate = 0
	}
	if commissionRate > 100 {
		commissionRate = 100
	}
	// 平台抽成 = 总金额 × 抽成比例 / 100，四舍五入到分
	commission := int64(math.Round(float64(totalCents) * commissionRate / 100))
	// 兜底：抽成不超过总金额
	if commission > totalCents {
		commission = totalCents
	}
	driverIncome := totalCents - commission
	return commission, driverIncome
}
