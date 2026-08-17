package rule

import (
	"errors"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
)

var (
	ErrRefundNotAllowed    = errors.New("only paid payment can be refunded")
	ErrRefundAmountInvalid = errors.New("refund amount must be positive")
	ErrRefundExceed        = errors.New("refund amount exceeds refundable amount")
)

// ValidateRefund 校验退款合法性。
//
// status        支付单当前状态
// amountCents   已支付金额（分）
// refundedCents 累计已退款金额（分）
// refundCents   本次申请退款金额（分）
func ValidateRefund(status int8, amountCents, refundedCents, refundCents int64) error {
	if status != model.PaymentStatusPaid {
		return ErrRefundNotAllowed
	}
	if refundCents <= 0 {
		return ErrRefundAmountInvalid
	}
	if refundedCents+refundCents > amountCents {
		return ErrRefundExceed
	}
	return nil
}
