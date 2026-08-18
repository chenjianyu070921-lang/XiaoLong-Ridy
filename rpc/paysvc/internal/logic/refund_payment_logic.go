package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RefundPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundPaymentLogic {
	return &RefundPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// genRefundNo 生成退款单号：RF + 时间戳 + 微秒。
func genRefundNo() string {
	return fmt.Sprintf("RF%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1000)
}

// 退款：校验并执行退款，回写支付单。
func (l *RefundPaymentLogic) RefundPayment(in *proto.RefundPaymentRequest) (*proto.RefundPaymentResponse, error) {
	repo := repository.NewPaymentRepo(l.svcCtx.DB)

	// 1. 查询支付单
	p, err := repo.FindByPaymentNo(l.ctx, in.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// 2. 校验退款合法性
	amountCents := priceutil.YuanToCents(p.Amount)
	refundedCents := priceutil.YuanToCents(p.RefundAmount)
	if err := rule.ValidateRefund(p.Status, amountCents, refundedCents, in.RefundAmountCents); err != nil {
		return nil, err
	}

	// 3. 调用渠道退款（本期 mock）
	refundNo := genRefundNo()
	ch := channel.NewMockChannel(p.Channel)
	if _, err := ch.Refund(l.ctx, p.PaymentNo, refundNo, in.RefundAmountCents); err != nil {
		return nil, err
	}

	// 4. 回写支付单
	totalRefunded := refundedCents + in.RefundAmountCents
	p.RefundAmount = priceutil.CentsToYuan(totalRefunded)
	// 全额退款才流转为「已退款」，部分退款保持「支付成功」
	if totalRefunded >= amountCents {
		p.Status = model.PaymentStatusRefund
	}
	if err := repo.Update(l.ctx, p); err != nil {
		return nil, err
	}

	return &proto.RefundPaymentResponse{
		Success:             true,
		RefundNo:            refundNo,
		RefundedAmountCents: totalRefunded,
	}, nil
}
