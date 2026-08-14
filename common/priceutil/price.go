// Package priceutil 提供金额单位换算与安全计算。
// 约定：接口与业务计算统一使用「分」(int64) 表示金额，避免浮点误差；
// 数据库存储使用 decimal(10,2) 的「元」(float64)。
package priceutil

import "math"

// YuanToCents 元转分：1.23 -> 123
func YuanToCents(yuan float64) int64 {
	return int64(math.Round(yuan * 100))
}

// CentsToYuan 分转元：123 -> 1.23
func CentsToYuan(cents int64) float64 {
	return float64(cents) / 100.0
}

// Add 金额相加（分），所有入参为分。
func Add(cents ...int64) int64 {
	var sum int64
	for _, c := range cents {
		sum += c
	}
	return sum
}

// Max 返回较大值，用于兜底非负金额。
func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
