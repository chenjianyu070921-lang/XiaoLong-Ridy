package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartAutoSettleJob 启动自动结算定时任务：
// 扫描已支付且未结算的支付单，按 defaultCommissionRate 调 SettleOrder 生成结算单。
func StartAutoSettleJob(ctx context.Context, svcCtx *svc.ServiceContext, interval time.Duration, defaultCommissionRate float64) {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	if defaultCommissionRate <= 0 {
		defaultCommissionRate = 20
	}
	payRepo := repository.NewPaymentRepo(svcCtx.DB)
	logx.Info("auto settle job started")
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logx.Info("auto settle job stopped")
				return
			case <-ticker.C:
				autoSettleOnce(ctx, svcCtx, payRepo, defaultCommissionRate)
			}
		}
	}()
}

func autoSettleOnce(ctx context.Context, svcCtx *svc.ServiceContext, payRepo *repository.PaymentRepo, rate float64) {
	runID := genRunID()
	settled := 0
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("auto settle panic: %v", r)
		}
		logx.Infof("auto settle run %s settled=%d", runID, settled)
	}()

	payments, err := payRepo.FindUnsettledPaidPayments(ctx, 200)
	if err != nil {
		logx.Errorf("find unsettled paid payments failed: %v", err)
		return
	}
	for _, p := range payments {
		driverID, err := svcCtx.OrderClient.GetDriverId(ctx, int64(p.OrderId))
		if err != nil {
			logx.Errorf("auto settle get driver_id for order %d failed: %v", p.OrderId, err)
			continue
		}
		l := NewSettleOrderLogic(ctx, svcCtx)
		_, err = l.SettleOrder(&proto.SettleOrderRequest{
			OrderId:         int64(p.OrderId),
			DriverId:        driverID,
			TotalAmountCents: p.AmountCents,
			CommissionRate:  rate,
		})
		if err != nil {
			logx.Errorf("auto settle order %d failed: %v", p.OrderId, err)
			continue
		}
		// 结算单已写入（SettleOrder 内事务）。给 settlement 打 auto_settled=1 + run_id 标记。
		if err := svcCtx.DB.WithContext(ctx).
			Model(&model.Settlement{}).
			Where("order_id = ?", p.OrderId).
			Updates(map[string]interface{}{
				"auto_settled":       1,
				"settled_job_run_id": runID,
			}).Error; err != nil {
			logx.Errorf("mark settlement auto_settled failed: %v", err)
		}
		settled++
	}
}


