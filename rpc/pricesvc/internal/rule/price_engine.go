// Package rule 提供计价引擎：根据计价规则与行程参数计算各分项费用。
// 所有金额统一使用「分」(int64)，避免浮点误差。
package rule

import (
	"errors"
	"math"

	"XiaoLong-Ridy/common/priceutil"
)

// PriceRuleInput 计价规则输入（金额均为分）。
type PriceRuleInput struct {
	BasePriceCents      int64   // 起步价（分）
	BaseDistanceM       int64   // 起步包含里程（米）
	PerKmPriceCents     int64   // 每公里价格（分）
	PerMinutePriceCents int64   // 每分钟时长费（分）
	NightSurchargeCents int64   // 夜间附加费（分/单）
	DynamicMaxFactor    float64 // 动态调价最大倍数
}

// PriceDetail 计价结果明细（金额均为分）。
type PriceDetail struct {
	BaseFeeCents      int64 // 起步价
	DistanceFeeCents  int64 // 里程费
	TimeFeeCents      int64 // 时长费
	NightFeeCents     int64 // 夜间附加费
	DynamicFeeCents   int64 // 动态调价费
	TotalCents        int64 // 合计
}

var (
	ErrNegativeDistance = errors.New("distance must be non-negative")
	ErrNegativeDuration = errors.New("duration must be non-negative")
)

// Estimate 计算一次行程的费用明细。
//
// distanceM  实际/预估里程（米）
// durationS  实际/预估时长（秒）
// isNight    是否处于夜间时段（由调用方根据规则时段判断）
// factor     动态调价倍数（1.0 表示不调价），受 DynamicMaxFactor 上限约束
func Estimate(r PriceRuleInput, distanceM, durationS int64, isNight bool, factor float64) (*PriceDetail, error) {
	if distanceM < 0 {
		return nil, ErrNegativeDistance
	}
	if durationS < 0 {
		return nil, ErrNegativeDuration
	}

	// 起步价
	baseFee := r.BasePriceCents

	// 里程费：超过起步里程部分按每公里计价（不足 1 公里按比例计）
	distanceFee := int64(0)
	if distanceM > r.BaseDistanceM {
		excessM := distanceM - r.BaseDistanceM
		distanceFee = excessM * r.PerKmPriceCents / 1000
	}

	// 时长费：按分钟计（不足 1 分钟按比例计）
	timeFee := durationS * r.PerMinutePriceCents / 60

	// 夜间附加费
	nightFee := int64(0)
	if isNight {
		nightFee = r.NightSurchargeCents
	}

	// 动态调价：基础价(起步+里程+时长) × (factor-1)，factor 不超过上限
	basic := priceutil.Add(baseFee, distanceFee, timeFee)
	cappedFactor := factor
	if cappedFactor > r.DynamicMaxFactor {
		cappedFactor = r.DynamicMaxFactor
	}
	if cappedFactor < 1.0 {
		cappedFactor = 1.0
	}
	dynamicFee := int64(0)
	if cappedFactor > 1.0 {
		// 动态费 = 基础价 * (factor - 1)，四舍五入到分
		dynamicFee = int64(math.Round(float64(basic) * (cappedFactor - 1.0)))
	}

	total := priceutil.Add(baseFee, distanceFee, timeFee, nightFee, dynamicFee)

	return &PriceDetail{
		BaseFeeCents:     baseFee,
		DistanceFeeCents: distanceFee,
		TimeFeeCents:     timeFee,
		NightFeeCents:    nightFee,
		DynamicFeeCents:  dynamicFee,
		TotalCents:       total,
	}, nil
}
