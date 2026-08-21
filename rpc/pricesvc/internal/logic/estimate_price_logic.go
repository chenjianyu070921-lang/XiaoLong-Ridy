package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/repository"
	"XiaoLong-Ridy/rpc/pricesvc/internal/rule"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

var ErrPriceRuleNotFound = errors.New("price rule not found")

type EstimatePriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEstimatePriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EstimatePriceLogic {
	return &EstimatePriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// EstimatePrice 行程价格预估：根据计价规则 + 里程/时长估算费用。
//
// 缺规则兜底（M5-12）：
//   - DB 查询不到任何启用的规则时，使用规则包内置的 DefaultPriceRule（系统兜底价）；
//   - 兜底规则显式标记 ID = DefaultPriceRuleID，便于后台对账识别；
//   - 不阻断流程：让乘客始终能看到估算价，对应兜底价由对账任务后续替换为正式规则。
func (l *EstimatePriceLogic) EstimatePrice(in *proto.EstimatePriceRequest) (*proto.EstimatePriceResponse, error) {
	ruleRepo := repository.NewPriceRuleRepo(l.svcCtx.DB)

	// 1. 查询启用的计价规则
	pr, err := ruleRepo.FindActive(l.ctx, in.CityCode, int8(in.CarType))
	usingDefault := false
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// M5-12：找不到规则时用系统兜底价，不阻断预估流程。
		l.Infof("price rule not found, fallback to default: city_code=%s, car_type=%d", in.CityCode, in.CarType)
		pr = rule.DefaultPriceRule(in.CityCode, int8(in.CarType))
		usingDefault = true
	}

	// 2. 判断是否夜间（根据规则时段 + 请求时刻）
	now := time.Now()
	if in.Timestamp > 0 {
		now = time.Unix(in.Timestamp, 0)
	}
	nightStart := ""
	nightEnd := ""
	if pr.NightStartTime != nil {
		nightStart = *pr.NightStartTime
	}
	if pr.NightEndTime != nil {
		nightEnd = *pr.NightEndTime
	}
	isNight := rule.IsNightTime(now, nightStart, nightEnd)

	// 3. 计算动态调价 factor：高峰时段自动上调
	factor := 1.0
	if rule.IsPeakTime(now) {
		factor = rule.PeakFactor
	}

	// 4. 调计价引擎计算费用明细
	detail, err := rule.Estimate(rule.PriceRuleInput{
		BasePriceCents:      priceutil.YuanToCents(pr.BasePrice),
		BaseDistanceM:       int64(pr.BaseDistanceKm * 1000),
		PerKmPriceCents:     priceutil.YuanToCents(pr.PerKmPrice),
		PerMinutePriceCents: priceutil.YuanToCents(pr.PerMinutePrice),
		NightSurchargeCents: priceutil.YuanToCents(pr.NightSurcharge),
		DynamicMaxFactor:    pr.DynamicMaxFactor,
	}, in.DistanceM, in.DurationS, isNight, factor)
	if err != nil {
		return nil, err
	}

	// 5. 响应里如果走的是兜底规则，price_rule_id 字段置 0（让调用方无法通过 ID 查到真实规则做误导），
	// 仅日志记录；这里保留 0 之外的 ID 仅便于后端对账。
	ruleId := int64(pr.Id)
	if usingDefault {
		ruleId = 0
	}

	return &proto.EstimatePriceResponse{
		PriceRuleId: ruleId,
		Detail: &proto.PriceDetail{
			BaseFeeCents:     detail.BaseFeeCents,
			DistanceFeeCents: detail.DistanceFeeCents,
			TimeFeeCents:     detail.TimeFeeCents,
			NightFeeCents:    detail.NightFeeCents,
			DynamicFeeCents:  detail.DynamicFeeCents,
			TotalCents:       detail.TotalCents,
		},
		TotalCents: detail.TotalCents,
	}, nil
}
