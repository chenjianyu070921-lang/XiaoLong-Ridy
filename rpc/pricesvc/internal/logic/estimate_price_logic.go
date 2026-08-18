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

// 行程价格预估：根据计价规则 + 里程/时长估算费用。
func (l *EstimatePriceLogic) EstimatePrice(in *proto.EstimatePriceRequest) (*proto.EstimatePriceResponse, error) {
	ruleRepo := repository.NewPriceRuleRepo(l.svcCtx.DB)

	// 1. 查询启用的计价规则
	pr, err := ruleRepo.FindActive(l.ctx, in.CityCode, int8(in.CarType))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPriceRuleNotFound
		}
		return nil, err
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

	return &proto.EstimatePriceResponse{
		PriceRuleId: int64(pr.Id),
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
