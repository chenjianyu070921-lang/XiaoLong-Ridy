package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartEventReconcileJob 启动支付成功事件对账补发任务。
// 定时扫描 status=paid 且 event_sent=false 的支付单，补发 Kafka 事件并标记 event_sent=true。
// 该任务作为 Kafka 发送失败后的兜底机制，避免"DB 已支付但事件未发送"造成下游订单状态不一致。
func StartEventReconcileJob(ctx context.Context, svcCtx *svc.ServiceContext, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	repo := repository.NewPaymentRepo(svcCtx.DB)
	logx.Info("payment event reconcile job started")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logx.Info("payment event reconcile job stopped")
				return
			case <-ticker.C:
				reconcileOnce(ctx, svcCtx, repo)
			}
		}
	}()
}

func reconcileOnce(ctx context.Context, svcCtx *svc.ServiceContext, repo *repository.PaymentRepo) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("event reconcile panic: %v", r)
		}
	}()

	payments, err := repo.FindUnsentPaidPayments(ctx, 100)
	if err != nil {
		logx.Errorf("find unsent paid payments failed: %v", err)
		return
	}
	if len(payments) == 0 {
		return
	}

	for _, p := range payments {
		if err := reconcilePayment(ctx, svcCtx, repo, p); err != nil {
			logx.Errorf("reconcile payment %s failed: %v", p.PaymentNo, err)
		}
	}
}

func reconcilePayment(ctx context.Context, svcCtx *svc.ServiceContext, repo *repository.PaymentRepo, p *model.Payment) error {
	// 使用默认 Unix 0 时间作为兜底，业务侧消费时自行判断。
	paidAt := int64(0)
	if p.PaidAt != nil {
		paidAt = p.PaidAt.Unix()
	}
	req := &proto.NotifyPaymentRequest{
		PaymentNo:     p.PaymentNo,
		TransactionId: p.TransactionId,
		TotalAmountCents: p.AmountCents,
		PaidAt:        paidAt,
	}

	logic := NewNotifyPaymentLogic(ctx, svcCtx)
	if err := logic.publishPaidEvent(req, p.AmountCents, int64(p.OrderId)); err != nil {
		return fmt.Errorf("publish paid event failed: %w", err)
	}
	if err := repo.UpdateSelective(ctx, p.Id, map[string]interface{}{"event_sent": true}); err != nil {
		return fmt.Errorf("mark event_sent=true failed: %w", err)
	}
	logx.Infof("reconcile payment %s event sent", p.PaymentNo)
	return nil
}
