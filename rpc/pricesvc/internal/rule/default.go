// Package rule 同时承载默认规则兜底，让 EstimatePrice 永远有可用规则。
// 当数据库无 price_rule 时使用 DefaultPriceRule，避免体验断点。
package rule

import (
	"math"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
)

// DefaultPriceRuleID 用于标记"系统兜底"使用的伪规则 ID。
//
// 选用 math.MaxInt64 作为 sentinel：
//   - 业务主键正常自增不会到达该值；
//   - sentinel 在响应里替换为 0 之前，能让"价格落库→后台对账"链路识别"兜底"分支；
//   - 选用大整数而非自造小常数（如 0xDEFA17），可降低与历史/迁移数据冲突的风险。
//
// 不要把这个值当作正常的 price_rule.id 写入；若要写入，请把对应 OrderPrice 行的 rule_id
// 置 0 并加一条对账标记（如 status=3 系统兜底）。
const DefaultPriceRuleID uint64 = math.MaxInt64

// DefaultPriceRule 返回兜底计价规则。
//
// 兜底规则按城市规模分级：
//   - 默认（city_code=空）：起步 10 元，包含 3 公里，每公里 2 元，每分钟 0.4 元，夜间附加 0；无调价上限。
//   - 后续可按 city_code 分级扩展（一线/二线/三线），本期按城市同价。
//
// 城市维度对兜底的影响通过 CityCode 字段留口，但本期不实现。
func DefaultPriceRule(cityCode string, carType int8) *model.PriceRule {
	_ = cityCode // 留口：未来按城市分级
	_ = carType  // 留口：未来按车型分级

	basePrice := priceutil.CentsToYuan(1000)              // 起步 10 元 = 1000 分
	perKmPrice := priceutil.CentsToYuan(200)              // 2 元/公里
	perMinutePrice := priceutil.CentsToYuan(40)           // 0.4 元/分钟
	nightSurcharge := priceutil.CentsToYuan(0)            // 兜底不收夜间费
	dynamicMax := 1.0                                     // 兜底不允许调价

	ns := "23:00:00"
	ne := "05:00:00"

	return &model.PriceRule{
		Id:               DefaultPriceRuleID,
		Name:             "系统兜底规则",
		CityCode:         cityCode,
		CarType:          carType,
		BasePrice:        basePrice,
		BaseDistanceKm:   3.00,
		PerKmPrice:       perKmPrice,
		PerMinutePrice:   perMinutePrice,
		NightStartTime:   &ns,
		NightEndTime:     &ne,
		NightSurcharge:   nightSurcharge,
		DynamicMaxFactor: dynamicMax,
		Status:           1, // 启用
	}
}

// IsUsingDefaultRule 判断 priceRuleId 是否使用了系统兜底规则（用于在响应里给前端提示）。
func IsUsingDefaultRule(priceRuleId int64) bool {
	return uint64(priceRuleId) == DefaultPriceRuleID
}
