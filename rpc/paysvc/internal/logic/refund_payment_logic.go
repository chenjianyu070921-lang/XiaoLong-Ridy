package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 退款流程幂等竞争错误：其它线程已经把状态改了，回滚事务并由上游判定。
var errRefundConcurrent = errors.New("refund concurrent update, please retry")

// newRefundRepo 抽出 repository 构造，避免 logic 直接依赖 repository 包（解耦便于测试替换）。
// 这里保留内联 repository.NewPaymentRepo 调用，方便后续切换测试桩。
var newRefundRepo = repository.NewPaymentRepo

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

// RefundPayment 退款：校验并执行退款，回写支付单。
//
// 事务边界：
//   - 校验 → 渠道退款 → 条件更新支付单，全程包在事务里；
//   - 仅当 status 仍为「支付成功」时才更新退款金额，杜绝并发退款重放；
//   - 渠道退款 RPC 仍放在事务外（同类原因：长耗时不该持 DB 连接）。
func (l *RefundPaymentLogic) RefundPayment(in *proto.RefundPaymentRequest) (*proto.RefundPaymentResponse, error) {
	// 1. 第一次读：校验状态、退款合法性，同时拿到支付单信息用于渠道退款入参。
	repo := newRefundRepo(l.svcCtx.DB)
	p, err := repo.FindByPaymentNo(l.ctx, in.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	amountCents := p.AmountCents
	refundedCents := p.RefundAmountCents
	if err := rule.ValidateRefund(p.Status, amountCents, refundedCents, in.RefundAmountCents); err != nil {
		return nil, err
	}

	// 2. 调渠道退款（事务外执行；失败直接回传，不进 DB）。
	refundNo := genRefundNo()
	ch := l.svcCtx.GetChannel(p.Channel)
	if _, err := ch.Refund(l.ctx, p.PaymentNo, refundNo, in.RefundAmountCents); err != nil {
		return nil, fmt.Errorf("channel refund failed: %w", err)
	}

	// 3. 条件更新 + 事务（M5-4 / M5-5）。
	totalRefunded := refundedCents + in.RefundAmountCents
	updates := map[string]interface{}{
		"refund_amount": totalRefunded,
	}
	// 全额退款才流转为「已退款」，部分退款保持「支付成功」。
	if totalRefunded >= amountCents {
		updates["status"] = model.PaymentStatusRefund
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Payment{}).
			Where("id = ? AND status = ?", p.Id, model.PaymentStatusPaid).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// 并发竞争：状态已变（被其它退款推进到「已退款」或被 Notify 改回 Pending 等）。
			return errRefundConcurrent
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &proto.RefundPaymentResponse{
		Success:             true,
		RefundNo:            refundNo,
		RefundedAmountCents: totalRefunded,
	}, nil
}
