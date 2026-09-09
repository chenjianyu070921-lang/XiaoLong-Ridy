package logic

import (
	"context"
	"encoding/json"
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

const refundIdempotencyTTL = 24 * time.Hour

var ErrRefundInProgress = errors.New("refund is processing")

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
	refundNo := in.GetRefundNo()
	if refundNo != "" && l.svcCtx.Redis != nil {
		key := refundIdempotencyKey(in.PaymentNo, refundNo)
		cached, err := l.svcCtx.Redis.Get(l.ctx, key).Bytes()
		if err == nil {
			var response proto.RefundPaymentResponse
			if json.Unmarshal(cached, &response) == nil {
				return &response, nil
			}
		}
		acquired, err := l.svcCtx.Redis.SetNX(l.ctx, key, "processing", refundIdempotencyTTL).Result()
		if err != nil {
			return nil, err
		}
		if !acquired {
			return nil, ErrRefundInProgress
		}
	}

	repo := repository.NewPaymentRepo(l.svcCtx.DB)

	// 1. 查询支付单
	p, err := repo.FindByPaymentNo(l.ctx, in.PaymentNo)
	if err != nil {
		l.releaseRefundIdempotency(in)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// 2. 校验退款合法性
	amountCents := priceutil.YuanToCents(p.Amount)
	refundedCents := priceutil.YuanToCents(p.RefundAmount)
	if err := rule.ValidateRefund(p.Status, amountCents, refundedCents, in.RefundAmountCents); err != nil {
		l.releaseRefundIdempotency(in)
		return nil, err
	}

	// 3. 调用渠道退款（本期 mock）。
	// 调用方传入退款单号时沿用该业务号，保证订单事件重试时具备稳定的退款标识；
	// 兼容未传退款单号的历史调用，仍由支付服务生成退款单号。
	if refundNo == "" {
		refundNo = genRefundNo()
	}
	ch := channel.NewMockChannel(p.Channel)
	if _, err := ch.Refund(l.ctx, p.PaymentNo, refundNo, in.RefundAmountCents); err != nil {
		l.releaseRefundIdempotency(in)
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
		l.releaseRefundIdempotency(in)
		return nil, err
	}

	response := &proto.RefundPaymentResponse{
		Success:             true,
		RefundNo:            refundNo,
		RefundedAmountCents: totalRefunded,
	}
	l.cacheRefundResponse(in, response)
	return response, nil
}

// refundIdempotencyKey 生成支付服务退款幂等键。
func refundIdempotencyKey(paymentNo, refundNo string) string {
	return "pay:refund:idem:" + paymentNo + ":" + refundNo
}

// releaseRefundIdempotency 释放失败请求的处理中标记，允许消息重新投递。
func (l *RefundPaymentLogic) releaseRefundIdempotency(in *proto.RefundPaymentRequest) {
	if l.svcCtx.Redis != nil && in.GetRefundNo() != "" {
		_ = l.svcCtx.Redis.Del(l.ctx, refundIdempotencyKey(in.PaymentNo, in.RefundNo)).Err()
	}
}

// cacheRefundResponse 缓存成功结果，重复事件直接返回原退款结果。
func (l *RefundPaymentLogic) cacheRefundResponse(in *proto.RefundPaymentRequest, response *proto.RefundPaymentResponse) {
	if l.svcCtx.Redis == nil || in.GetRefundNo() == "" {
		return
	}
	data, err := json.Marshal(response)
	if err != nil {
		return
	}
	if err := l.svcCtx.Redis.Set(l.ctx, refundIdempotencyKey(in.PaymentNo, in.RefundNo), data, refundIdempotencyTTL).Err(); err != nil {
		l.Logger.Errorf("cache refund response failed, paymentNo=%s refundNo=%s: %v", in.PaymentNo, in.RefundNo, err)
	}
}
