package rule

import (
	"errors"
)

// 优惠券类型
const (
	CouponTypeFixed    = 1 // 固定金额券
	CouponTypeDiscount = 2 // 折扣券
)

// CouponInput 优惠券信息（金额均为分）。
type CouponInput struct {
	Type            int32 // 1固定金额 2折扣
	FaceValueCents  int64 // 固定金额券面额（分）
	Discount        int32 // 折扣券折扣，80 表示 8 折
	ThresholdCents  int64 // 使用门槛（分），0 表示无门槛
	MaxDiscountCents int64 // 折扣券最大优惠金额（分），0 表示不限制
}

var (
	ErrCouponNotMeetThreshold = errors.New("order amount does not meet coupon threshold")
)

// DiscountResult 优惠计算结果（金额均为分）。
type DiscountResult struct {
	DiscountAmountCents  int64 // 优惠券抵扣金额
	PlatformSubsidyCents int64 // 平台补贴金额（本期简化为 0）
	PayableAmountCents   int64 // 乘客实付金额
}

// CalculateDiscount 计算优惠券抵扣。
// totalCents 为抵扣前订单总金额（分）。
func CalculateDiscount(totalCents int64, c CouponInput) (*DiscountResult, error) {
	// 门槛校验
	if c.ThresholdCents > 0 && totalCents < c.ThresholdCents {
		return nil, ErrCouponNotMeetThreshold
	}

	var discountCents int64

	switch c.Type {
	case CouponTypeFixed:
		// 固定金额券：直接减面额，不超过订单金额
		discountCents = c.FaceValueCents
		if discountCents > totalCents {
			discountCents = totalCents
		}
	case CouponTypeDiscount:
		// 折扣券：优惠 = 订单金额 * (1 - 折扣/100)
		// 例：8 折（discount=80）→ 优惠 20%
		if c.Discount <= 0 || c.Discount >= 100 {
			discountCents = 0
		} else {
			discountCents = totalCents - totalCents*int64(c.Discount)/100
			// 受最大优惠金额约束
			if c.MaxDiscountCents > 0 && discountCents > c.MaxDiscountCents {
				discountCents = c.MaxDiscountCents
			}
		}
	default:
		discountCents = 0
	}

	// 兜底：优惠不超过订单金额
	if discountCents < 0 {
		discountCents = 0
	}
	if discountCents > totalCents {
		discountCents = totalCents
	}

	payable := totalCents - discountCents

	return &DiscountResult{
		DiscountAmountCents:  discountCents,
		PlatformSubsidyCents: 0, // 本期平台补贴为 0，后续可扩展
		PayableAmountCents:   payable,
	}, nil
}

// 辅助：判断订单是否满足门槛（供 logic 层提前校验用）。
func MeetThreshold(totalCents int64, c CouponInput) bool {
	if c.ThresholdCents <= 0 {
		return true
	}
	return totalCents >= c.ThresholdCents
}
