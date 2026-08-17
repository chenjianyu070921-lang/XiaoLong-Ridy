package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/pricesvc/internal/repository"
	"XiaoLong-Ridy/rpc/pricesvc/internal/rule"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CalculateDiscountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCalculateDiscountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateDiscountLogic {
	return &CalculateDiscountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 优惠券抵扣计算：根据优惠券计算折扣与实付金额。
func (l *CalculateDiscountLogic) CalculateDiscount(in *proto.CalculateDiscountRequest) (*proto.CalculateDiscountResponse, error) {
	// 1. 构造优惠券输入
	coupon := rule.CouponInput{
		Type:             int32(in.Coupon.Type),
		FaceValueCents:   in.Coupon.FaceValueCents,
		Discount:         in.Coupon.Discount,
		ThresholdCents:   in.Coupon.ThresholdCents,
		MaxDiscountCents: in.Coupon.MaxDiscountCents,
	}

	// 2. 计算优惠
	res, err := rule.CalculateDiscount(in.TotalCents, coupon)
	if err != nil {
		if errors.Is(err, rule.ErrCouponNotMeetThreshold) {
			return nil, err
		}
		return nil, err
	}

	// 3. 更新 order_price（若存在）
	priceRepo := repository.NewOrderPriceRepo(l.svcCtx.DB)
	op, err := priceRepo.FindByOrderId(l.ctx, uint64(in.OrderId))
	if err == nil {
		op.DiscountAmount = priceutil.CentsToYuan(res.DiscountAmountCents)
		op.PlatformSubsidy = priceutil.CentsToYuan(res.PlatformSubsidyCents)
		op.PayableAmount = priceutil.CentsToYuan(res.PayableAmountCents)
		op.Status = 2 // 已确认
		if err := priceRepo.Update(l.ctx, op); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// 若 order_price 不存在则仅计算返回，不落库（预估阶段可能尚未生成明细）

	return &proto.CalculateDiscountResponse{
		DiscountAmountCents:  res.DiscountAmountCents,
		PlatformSubsidyCents: res.PlatformSubsidyCents,
		PayableAmountCents:   res.PayableAmountCents,
	}, nil
}
