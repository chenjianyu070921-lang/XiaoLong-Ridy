package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ChannelTransaction 渠道侧支付快照（用于对账）。真实由支付宝对账接口填充。
type ChannelTransaction struct {
	TransactionId string // 渠道流水号
	AmountCents   int64  // 渠道金额（分）
	Status        string // 渠道侧状态
}

// QueryChannelTransactions 拉取渠道侧支付流水快照，按 payment_no 建索引。
//
// 本函数作为对账对接点：结构骨架阶段返回空 map，差异检测天然走"渠道无"分支，
// 供本地联调验证整条链路；生产需接入支付宝 alipay.trade.query 批量查询。
var QueryChannelTransactions = func(ctx context.Context, paymentNos []string) (map[string]*ChannelTransaction, error) {
	// TODO(生产): 调用支付宝 alipay.trade.query 批量查询，按 out_trade_no 建立索引。
	return map[string]*ChannelTransaction{}, nil
}

// StartPaymentReconcileJob 启动支付渠道对账定时任务。
//
// 扫描最近 lookback 窗口内已支付的支付单，拉取渠道快照做差异比对，
// 差异写 payment_reconcile_diff，执行过程写 payment_channel_reconcile_log。
func StartPaymentReconcileJob(ctx context.Context, svcCtx *svc.ServiceContext, interval, lookback time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if lookback <= 0 {
		lookback = 30 * time.Minute
	}
	repo := repository.NewPaymentReconcileRepo(svcCtx.DB)
	logx.Info("payment reconcile job started")
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logx.Info("payment reconcile job stopped")
				return
			case <-ticker.C:
				channelReconcileOnce(ctx, svcCtx, repo, lookback)
			}
		}
	}()
}

func channelReconcileOnce(ctx context.Context, svcCtx *svc.ServiceContext, repo *repository.PaymentReconcileRepo, lookback time.Duration) {
	runID := genRunID()
	runLog := &model.PaymentChannelReconcileLog{
		RunId:     runID,
		StartedAt: time.Now(),
		Status:    model.ReconcileLogRunning,
	}
	if err := repo.CreateRunLog(ctx, runLog); err != nil {
		logx.Errorf("create reconcile run log failed: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			_ = repo.FinishRunLog(ctx, runID, 0, 0, model.ReconcileLogFailed, fmtRecover(r))
			logx.Errorf("reconcile panic: %v", r)
		}
	}()

	since := time.Now().Add(-lookback)
	payments, err := repo.ListPaidPaymentsForReconcile(ctx, since, 500)
	if err != nil {
		_ = repo.FinishRunLog(ctx, runID, 0, 0, model.ReconcileLogFailed, err.Error())
		logx.Errorf("list paid payments for reconcile failed: %v", err)
		return
	}

	channelMap, err := QueryChannelTransactions(ctx, paymentNosOf(payments))
	if err != nil {
		_ = repo.FinishRunLog(ctx, runID, 0, 0, model.ReconcileLogFailed, err.Error())
		logx.Errorf("query channel transactions failed: %v", err)
		return
	}

	diffCount := 0
	for _, p := range payments {
		if diff := detectDiff(p, channelMap[p.PaymentNo]); diff != nil {
			diff.RunId = runID
			if err := repo.InsertDiff(ctx, diff); err != nil {
				logx.Errorf("insert reconcile diff failed: %v", err)
				continue
			}
			diffCount++
		}
	}

	_ = repo.FinishRunLog(ctx, runID, len(payments), diffCount, model.ReconcileLogSuccess, "")
	logx.Infof("reconcile run %s done: scanned=%d diff=%d", runID, len(payments), diffCount)
}

// detectDiff 比对平台支付单与渠道侧快照，返回差异（nil 表示无差异）。
func detectDiff(p *model.Payment, ch *ChannelTransaction) *model.PaymentReconcileDiff {
	if ch == nil {
		// 平台已支付但渠道查不到 → 记为渠道缺失差异。
		return &model.PaymentReconcileDiff{
			PaymentNo:      p.PaymentNo,
			OrderId:        p.OrderId,
			DiffType:       model.ReconcileDiffChannelOnly,
			PlatformAmount: p.AmountCents,
			ChannelAmount:  0,
			PlatformStatus: int8(p.Status),
			ChannelStatus:  "",
			ChannelTxId:    "",
			DetectedAt:     time.Now(),
			Remark:         "channel transaction not found, pending alipay reconcile api integration",
		}
	}
	// 有渠道快照：金额不一致 → 差异。
	if ch.AmountCents != p.AmountCents {
		return &model.PaymentReconcileDiff{
			PaymentNo:      p.PaymentNo,
			OrderId:        p.OrderId,
			DiffType:       model.ReconcileDiffAmount,
			PlatformAmount: p.AmountCents,
			ChannelAmount:  ch.AmountCents,
			PlatformStatus: int8(p.Status),
			ChannelStatus:  ch.Status,
			ChannelTxId:    ch.TransactionId,
			DetectedAt:     time.Now(),
			Remark:         "amount mismatch",
		}
	}
	return nil
}

func paymentNosOf(payments []*model.Payment) []string {
	out := make([]string, 0, len(payments))
	for _, p := range payments {
		out = append(out, p.PaymentNo)
	}
	return out
}

// genRunID 生成对账执行 ID：时间戳 + 随机 hex。
func genRunID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return time.Now().Format("20060102150405") + "-" + hex.EncodeToString(b)
}

func fmtRecover(r interface{}) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return "panic"
}
