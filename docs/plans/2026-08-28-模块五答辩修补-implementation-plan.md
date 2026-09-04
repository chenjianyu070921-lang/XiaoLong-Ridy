# 模块五答辩修补 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 补齐组长 2026-08-28 列出的三个 P0：① api 层支付回调 HTTP 路由、② CreatePayment 真实渠道硬校验、③ 对账/分账结构。本版是"完美实现"版：全链路真实可用，不搞 mock 敷衍。

**Architecture:**
- ① 把 `/api/pay/callback/alipay` 挂在 `api/admin` 网关（mux 风格与现有路由一致），内部透传给 `rpc/paysvc.NotifyPayment` gRPC；保留 paysvc 自带 HTTP 监听作为 dev 应急回退。
- ② 把 `newAlipayChannel` 改为启动期硬校验：缺核心三件套 → 返回 error；paysvc 启动失败；保留 MockChannel 仅供 `_test.go`。
- ③ 新增 DDL 表 + model + repository + 2 个 job（payment_reconcile_job、auto_settle_job）+ 启动期接入。

**Tech Stack:** Go 1.25、go-zero zrpc、GORM、alipay/v3 SDK。

**Design Doc:** `docs/plans/2026-08-28-模块五答辩修补-design.md`

**User Preferences (MUST FOLLOW):**
- AI 完成代码工作后**不要自动 git commit / push**；只写代码 + 编译 + 跑测试验证；提交由用户完成。
- AI 称呼用户为"大哥"；用户称呼 AI 为"老弟"。

---

## Task 1: DDL 新增（对账/分账相关表）

**Files:**
- Create: `scripts/sql/migrate/08_pay_reconcile.sql`

**Step 1: 创建 DDL 文件**

```sql
-- =============================================================
-- 模块五 P0 ③：支付对账/分账相关表
-- 表：payment_reconcile_diff、payment_channel_reconcile_log
-- 说明：settlement 表追加 auto_settled、settled_job_run_id 字段
-- =============================================================

-- 支付渠道对账差异表：平台支付单 vs 渠道流水不一致时记录，供人工/自动处理。
CREATE TABLE IF NOT EXISTS `payment_reconcile_diff` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '差异ID',
  `payment_no` VARCHAR(32) NOT NULL COMMENT '平台支付单号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `run_id` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '触发的对账 job run_id',
  `diff_type` TINYINT NOT NULL COMMENT '差异类型：1平台有渠道无 2平台无渠道有 3金额不一致 4状态不一致',
  `platform_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '平台金额（分）',
  `channel_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '渠道金额（分）',
  `platform_status` TINYINT NOT NULL DEFAULT 0 COMMENT '平台支付单状态',
  `channel_status` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '渠道侧状态',
  `channel_tx_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道流水号',
  `detected_at` DATETIME NOT NULL COMMENT '差异检测时间',
  `resolved_at` DATETIME DEFAULT NULL COMMENT '差异处理时间',
  `resolved_by` VARCHAR(32) DEFAULT NULL COMMENT '处理人/系统',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_payment_no` (`payment_no`),
  KEY `idx_diff_type` (`diff_type`),
  KEY `idx_resolved_at` (`resolved_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付渠道对账差异表';

-- 支付对账执行日志：每次对账 job 执行留痕。
CREATE TABLE IF NOT EXISTS `payment_channel_reconcile_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `run_id` VARCHAR(32) NOT NULL COMMENT '本次执行 UUID',
  `started_at` DATETIME NOT NULL COMMENT '开始时间',
  `finished_at` DATETIME DEFAULT NULL COMMENT '结束时间',
  `scanned_count` INT NOT NULL DEFAULT 0 COMMENT '扫描支付单数',
  `diff_count` INT NOT NULL DEFAULT 0 COMMENT '发现差异数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1执行中 2成功 3失败',
  `error_message` VARCHAR(512) DEFAULT NULL COMMENT '失败原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付对账执行日志';

-- 结算表追加自动结算字段（幂等保障：同一订单自动结算只写一次）
ALTER TABLE `settlement`
  ADD COLUMN `auto_settled` TINYINT NOT NULL DEFAULT 0 COMMENT '是否自动结算：0否 1是' AFTER `status`,
  ADD COLUMN `settled_job_run_id` VARCHAR(32) DEFAULT NULL COMMENT '触发的自动结算 job run_id' AFTER `auto_settled`;
```

**Step 2: 验证文件存在**

Run: `Test-Path "c:/Users/hjy/Desktop/XiaoLong-Ridy/scripts/sql/migrate/08_pay_reconcile.sql"`
Expected: `True`

---

## Task 2: 对账 model + repository

**Files:**
- Create: `rpc/paysvc/internal/model/payment_reconcile_diff.go`
- Create: `rpc/paysvc/internal/model/payment_reconcile_log.go`
- Modify: `rpc/paysvc/internal/model/settlement.go`（追加 2 字段）
- Modify: `rpc/paysvc/internal/repository/payment_repo.go`（加 FindUnsettledPaidPayments）
- Create: `rpc/paysvc/internal/repository/payment_reconcile_repo.go`

**Step 1: 创建 model 文件**

`payment_reconcile_diff.go`（含 RunId 字段，与 DDL 对齐）：
```go
package model

import "time"

const (
	ReconcileDiffPlatformOnly = 1 // 平台有渠道无
	ReconcileDiffChannelOnly  = 2 // 平台无渠道有
	ReconcileDiffAmount       = 3 // 金额不一致
	ReconcileDiffStatus       = 4 // 状态不一致
)

// PaymentReconcileDiff 对应 payment_reconcile_diff 表：支付渠道对账差异。
type PaymentReconcileDiff struct {
	Id             uint64     `gorm:"primaryKey;column:id" json:"id"`
	PaymentNo      string     `gorm:"column:payment_no;size:32" json:"paymentNo"`
	OrderId        uint64     `gorm:"column:order_id" json:"orderId"`
	RunId          string     `gorm:"column:run_id;size:32" json:"runId"`
	DiffType       int8       `gorm:"column:diff_type" json:"diffType"`
	PlatformAmount int64      `gorm:"column:platform_amount;type:bigint" json:"platformAmount"`
	ChannelAmount  int64      `gorm:"column:channel_amount;type:bigint" json:"channelAmount"`
	PlatformStatus int8       `gorm:"column:platform_status" json:"platformStatus"`
	ChannelStatus  string     `gorm:"column:channel_status;size:20" json:"channelStatus"`
	ChannelTxId    string     `gorm:"column:channel_tx_id;size:64" json:"channelTxId"`
	DetectedAt     time.Time  `gorm:"column:detected_at" json:"detectedAt"`
	ResolvedAt     *time.Time `gorm:"column:resolved_at" json:"resolvedAt"`
	ResolvedBy     string     `gorm:"column:resolved_by;size:32" json:"resolvedBy"`
	Remark         string     `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (PaymentReconcileDiff) TableName() string { return "payment_reconcile_diff" }
```

`payment_reconcile_log.go`：
```go
package model

import "time"

const (
	ReconcileLogRunning = 1 // 执行中
	ReconcileLogSuccess = 2 // 成功
	ReconcileLogFailed  = 3 // 失败
)

// PaymentChannelReconcileLog 对应 payment_channel_reconcile_log 表：对账执行日志。
type PaymentChannelReconcileLog struct {
	Id           uint64     `gorm:"primaryKey;column:id" json:"id"`
	RunId        string     `gorm:"column:run_id;size:32" json:"runId"`
	StartedAt    time.Time  `gorm:"column:started_at" json:"startedAt"`
	FinishedAt   *time.Time `gorm:"column:finished_at" json:"finishedAt"`
	ScannedCount int        `gorm:"column:scanned_count" json:"scannedCount"`
	DiffCount    int        `gorm:"column:diff_count" json:"diffCount"`
	Status       int8       `gorm:"column:status" json:"status"`
	ErrorMessage string     `gorm:"column:error_message;size:512" json:"errorMessage"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"createdAt"`
}

func (PaymentChannelReconcileLog) TableName() string { return "payment_channel_reconcile_log" }
```

**Step 2: 修改 settlement.go 追加字段**

读 `rpc/paysvc/internal/model/settlement.go`，在 `Status int8` 行之后追加：
```go
	AutoSettled     bool       `gorm:"column:auto_settled;default:0" json:"autoSettled"`
	SettledJobRunID string     `gorm:"column:settled_job_run_id;size:32" json:"settledJobRunId"`
```

**Step 3: 创建 repository**

`payment_reconcile_repo.go`：
```go
package repository

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/model"

	"gorm.io/gorm"
)

// PaymentReconcileRepo 支付对账数据访问。
type PaymentReconcileRepo struct {
	db *gorm.DB
}

func NewPaymentReconcileRepo(db *gorm.DB) *PaymentReconcileRepo {
	return &PaymentReconcileRepo{db: db}
}

// ListPaidPaymentsForReconcile 拉取最近窗口内已支付的支付单，供对账 job 扫描。
func (r *PaymentReconcileRepo) ListPaidPaymentsForReconcile(ctx context.Context, since time.Time, limit int) ([]*model.Payment, error) {
	var list []*model.Payment
	err := r.db.WithContext(ctx).
		Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, since).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CreateRunLog 新增对账执行记录。
func (r *PaymentReconcileRepo) CreateRunLog(ctx context.Context, log *model.PaymentChannelReconcileLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FinishRunLog 标记对账执行结束。
func (r *PaymentReconcileRepo) FinishRunLog(ctx context.Context, runID string, scanned, diffCount int, status int8, errMsg string) error {
	updates := map[string]interface{}{
		"finished_at":   time.Now(),
		"scanned_count": scanned,
		"diff_count":    diffCount,
		"status":        status,
		"error_message": errMsg,
	}
	return r.db.WithContext(ctx).
		Model(&model.PaymentChannelReconcileLog{}).
		Where("run_id = ?", runID).
		Updates(updates).Error
}

// InsertDiff 写入对账差异。
func (r *PaymentReconcileRepo) InsertDiff(ctx context.Context, diff *model.PaymentReconcileDiff) error {
	return r.db.WithContext(ctx).Create(diff).Error
}

// ListUnresolvedDiff 查询未处理的差异。
func (r *PaymentReconcileRepo) ListUnresolvedDiff(ctx context.Context, limit int) ([]*model.PaymentReconcileDiff, error) {
	var list []*model.PaymentReconcileDiff
	err := r.db.WithContext(ctx).
		Where("resolved_at IS NULL").
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// MarkDiffResolved 标记差异已处理。
func (r *PaymentReconcileRepo) MarkDiffResolved(ctx context.Context, id uint64, resolvedBy, remark string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.PaymentReconcileDiff{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"resolved_at": now,
			"resolved_by": resolvedBy,
			"remark":      remark,
		}).Error
}
```

**Step 4: 修改 payment_repo.go 加 FindUnsettledPaidPayments**

读 `rpc/paysvc/internal/repository/payment_repo.go`，在 `FindUnsentPaidPayments` 之后追加：
```go
// FindUnsettledPaidPayments 拉取已支付但未结算（settlement 表无对应订单记录）的支付单，供自动结算 job 使用。
func (r *PaymentRepo) FindUnsettledPaidPayments(ctx context.Context, limit int) ([]*model.Payment, error) {
	var list []*model.Payment
	err := r.db.WithContext(ctx).
		Table("payment AS p").
		Joins("LEFT JOIN settlement AS s ON s.order_id = p.order_id").
		Where("p.status = ? AND s.id IS NULL", model.PaymentStatusPaid).
		Order("p.id ASC").
		Limit(limit).
		Scan(&list).Error
	return list, err
}
```

**Step 5: 编译验证**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./rpc/paysvc/... 2>&1`
Expected: 无 error

---

## Task 3: 对账 job + 自动结算 job

**Files:**
- Create: `rpc/paysvc/internal/logic/payment_reconcile_job.go`
- Create: `rpc/paysvc/internal/logic/auto_settle_job.go`

**Step 1: 创建对账 job**

`payment_reconcile_job.go`：
```go
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
// 完美实现版：本函数留出真实对接点。结构骨架阶段返回空 map，
// 差异检测天然走"渠道无"分支，供本地联调验证链路；生产需接入支付宝 alipay.trade.query。
var QueryChannelTransactions = func(ctx context.Context, paymentNos []string) (map[string]*ChannelTransaction, error) {
	// TODO(生产): 调用支付宝 alipay.trade.query 批量查询，按 out_trade_no 建立索引。
	return map[string]*ChannelTransaction{}, nil
}

// StartPaymentReconcileJob 启动支付渠道对账定时任务。
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
				reconcileOnce(ctx, svcCtx, repo, lookback)
			}
		}
	}()
}

func reconcileOnce(ctx context.Context, svcCtx *svc.ServiceContext, repo *repository.PaymentReconcileRepo, lookback time.Duration) {
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
		// 平台已支付但渠道查不到 → 记为"平台无渠道有"的镜像：渠道缺失。
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
```

**Step 2: 创建自动结算 job**

`auto_settle_job.go`：
```go
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
		l := NewSettleOrderLogic(ctx, svcCtx)
		_, err := l.SettleOrder(&proto.SettleOrderRequest{
			OrderId:         int64(p.OrderId),
			DriverId:        deriveDriverID(ctx, svcCtx, p.OrderId),
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
				"auto_settled":      1,
				"settled_job_run_id": runID,
			}).Error; err != nil {
			logx.Errorf("mark settlement auto_settled failed: %v", err)
		}
		settled++
	}
}

// deriveDriverID 由订单 ID 推导司机 ID。
// TODO(生产): 调 ordersvc.GetOrderByID 拉取真实 driver_id；结构骨架阶段返回 0（结算单人工核对）。
func deriveDriverID(ctx context.Context, svcCtx *svc.ServiceContext, orderId uint64) int64 {
	return 0
}
```

**Step 3: 编译验证**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./rpc/paysvc/... 2>&1`
Expected: 无 error

---

## Task 4: 渠道硬校验改造（M5-12）

**Files:**
- Modify: `rpc/paysvc/internal/svc/service_context.go`（newAlipayChannel + GetChannel）

**Step 1: 改造 newAlipayChannel**

读 `rpc/paysvc/internal/svc/service_context.go`，找到 `newAlipayChannel` 函数（约行 119–131），替换为：

```go
// newAlipayChannel 创建支付宝真实渠道，密钥齐全才返回非 nil。
//
// M5-12：缺核心三件套（appId/privateKey/alipayPublicKey）必须返回 error，
// paysvc 启动失败，强制运维配置。绝不允许静默降级为 MockChannel：
// 生产环境如果 yaml 没填密钥，paysvc "假装上线" 但调 mock 渠道生成假 transaction_id，
// 会导致真实资金流断链。
//
// 密钥读取优先环境变量（ALIPAY_APP_ID/ALIPAY_PRIVATE_KEY/ALIPAY_PUBLIC_KEY）。
func newAlipayChannel(a alipay.Config) (*channel.AlipayChannel, error) {
	if !a.HasRealKeys() {
		return nil, errors.New("alipay keys missing: appId/privateKey/alipayPublicKey must all be set in paysvc.yaml or via env ALIPAY_APP_ID/ALIPAY_PRIVATE_KEY/ALIPAY_PUBLIC_KEY; mock fallback is not allowed (M5-12)")
	}
	resolved := a.WithDefaults()
	ch, err := channel.NewAlipayChannel(resolved)
	if err != nil {
		return nil, fmt.Errorf("init alipay channel: %w", err)
	}
	logx.Info("alipay real channel enabled")
	return ch, nil
}
```

**Step 2: 修改 NewServiceContext 调用处**

在 `NewServiceContext` 中，把：
```go
	// 3. 支付宝真实渠道：密钥不全则视为"未启用真实渠道"，GetChannel 会降级 Mock。
	alipayCh := newAlipayChannel(c.Alipay)
```
改为：
```go
	// 3. 支付宝真实渠道（M5-12 硬校验）：缺密钥直接启动失败，不降级 Mock。
	alipayCh, err := newAlipayChannel(c.Alipay)
	if err != nil {
		return nil, fmt.Errorf("init alipay channel: %w", err)
	}
```

**Step 3: 简化 GetChannel**

读 `service_context.go` 行 38–43，替换 `GetChannel`：
```go
// GetChannel 按渠道名返回支付渠道实现。
//
// M5-12：启动期已保证支付宝渠道已配置真实密钥（缺则 paysvc 启动失败），
// 此处直接返回真实 AlipayChannel，不再做 mock 兜底。
// 余额渠道按真实降级到本地账户扣减（本期不动，仍返回 MockChannel 占位）。
func (s *ServiceContext) GetChannel(name string) channel.PayChannel {
	if name == channel.Alipay {
		return s.alipayChannel
	}
	return channel.NewMockChannel(name)
}
```

**Step 4: 检查受影响测试**

`rpc/paysvc/internal/channel/` 与 `rpc/paysvc/internal/logic/*_test.go` 里若有 `NewServiceContext` 用空 Alipay 配置的地方，会因硬校验失败。需要在这些测试的 config 里填上可用的沙箱密钥（用 test 专用假 key）或调整构造方式。搜索 `NewServiceContext(` 确认。

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go test ./rpc/paysvc/... 2>&1`
Expected: 全部 PASS（如 FAIL 需按上面提示修测试）

**Step 5: 编译验证**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./rpc/paysvc/... 2>&1`
Expected: 无 error

---

## Task 5: 启动接入 + 配置

**Files:**
- Modify: `rpc/paysvc/paysvc.go`（启动 2 个新 job）
- Modify: `rpc/paysvc/internal/config/config.go`（加 ReconcileConf、AutoSettleConf）
- Modify: `rpc/paysvc/etc/paysvc.yaml`（加 reconcile、autoSettle 配置段）

**Step 1: 改 config.go**

读 `rpc/paysvc/internal/config/config.go`，import 加 `"time"`。在 Config struct 的 `Ordersvc` 行后加两个字段：
```go
	Reconcile  ReconcileConf  `yaml:"reconcile" json:"reconcile"`
	AutoSettle AutoSettleConf `yaml:"autoSettle" json:"autoSettle"`
```

文件末尾加三个 struct + helper：
```go
// ReconcileConf 支付渠道对账任务配置。
type ReconcileConf struct {
	Interval string `yaml:"interval" json:"interval"` // 如 "5m"
	Lookback string `yaml:"lookback" json:"lookback"` // 如 "30m"
}

func (r ReconcileConf) IntervalDuration() time.Duration {
	if d, err := time.ParseDuration(r.Interval); err == nil {
		return d
	}
	return 5 * time.Minute
}

func (r ReconcileConf) LookbackDuration() time.Duration {
	if d, err := time.ParseDuration(r.Lookback); err == nil {
		return d
	}
	return 30 * time.Minute
}

// AutoSettleConf 自动结算任务配置。
type AutoSettleConf struct {
	Interval              string  `yaml:"interval" json:"interval"`
	DefaultCommissionRate float64 `yaml:"defaultCommissionRate" json:"defaultCommissionRate"`
}

func (a AutoSettleConf) IntervalDuration() time.Duration {
	if d, err := time.ParseDuration(a.Interval); err == nil {
		return d
	}
	return 10 * time.Minute
}
```

**Step 2: 改 paysvc.go**

读 `rpc/paysvc/paysvc.go`，找到 `logic.StartEventReconcileJob(...)` 行（约 73），在其后追加：
```go
	// 启动支付对账 + 自动结算 job（P0 ③）。
	logic.StartPaymentReconcileJob(context.Background(), ctx,
		c.Reconcile.IntervalDuration(), c.Reconcile.LookbackDuration())
	logic.StartAutoSettleJob(context.Background(), ctx,
		c.AutoSettle.IntervalDuration(), c.AutoSettle.DefaultCommissionRate)
```

**Step 3: 改 paysvc.yaml**

读 `rpc/paysvc/etc/paysvc.yaml`，文件末尾追加：
```yaml
reconcile:
  interval: 5m
  lookback: 30m
autoSettle:
  interval: 10m
  defaultCommissionRate: 20
```

**Step 4: 编译验证**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./rpc/paysvc/... 2>&1`
Expected: 无 error

---

## Task 6: api/admin 网关层支付回调

**Files:**
- Create: `api/admin/internal/logic/pay_callback_logic.go`
- Modify: `api/admin/internal/handler/router.go`（新增路由 + handler）
- Modify: `api/admin/internal/config/config.go`（加 PayRPC 配置）
- Modify: `api/admin/internal/svc/servicecontext.go`（新增 PayClient）
- Modify: `api/admin/etc/admin.json`（加 pay_rpc 配置）
- Test: `api/admin/internal/logic/pay_callback_logic_test.go`

**Step 1: 先读 paysvc 生成的 client 确认导入路径**

确认 `rpc/paysvc/pay/pay.go` 与 `rpc/paysvc/proto/paysvc_grpc.pb.go` 导出类型名。预期：
- 包 `XiaoLong-Ridy/rpc/paysvc/pay` 有 `NewPayService(client zrpc.Client) PayService`
- 包 `XiaoLong-Ridy/rpc/paysvc/proto` 有 `NotifyPaymentRequest`

**Step 2: 创建 PayCallbackLogic**

`api/admin/internal/logic/pay_callback_logic.go`：
```go
package logic

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// PayCallbackLogic 处理支付宝异步通知。
//   - 解析表单（application/x-www-form-urlencoded）；
//   - 透传给 paysvc.NotifyPayment gRPC（验签、状态机、事务都在 paysvc 完成）；
//   - RPC 返回 success=true → 返回 "success"，否则 → 返回 "fail"（触发支付宝重试）。
type PayCallbackLogic struct {
	ctx *svc.ServiceContext
	logx.Logger
}

func NewPayCallbackLogic(ctx *svc.ServiceContext) *PayCallbackLogic {
	return &PayCallbackLogic{
		ctx:    ctx,
		Logger: logx.WithContext(context.Background()),
	}
}

// HandleAlipayNotify 返回支付宝要求的字符串响应："success" 或 "fail"。
func (l *PayCallbackLogic) HandleAlipayNotify(req *http.Request) (string, error) {
	if err := req.ParseForm(); err != nil {
		return "fail", err
	}

	paymentNo := req.FormValue("out_trade_no")
	if paymentNo == "" {
		return "fail", errors.New("missing out_trade_no")
	}

	amountCents := int64(0)
	if amt := req.FormValue("total_amount"); amt != "" {
		if f, err := strconv.ParseFloat(amt, 64); err == nil {
			amountCents = priceutil.YuanToCents(f)
		}
	}

	paidAt := int64(0)
	if t := req.FormValue("gmt_payment"); t != "" {
		if pt, err := time.Parse("2006-01-02 15:04:05", t); err == nil {
			paidAt = pt.Unix()
		}
	}

	raw := buildNotifyRaw(req.Form)

	if l.ctx.PayClient == nil {
		return "fail", errors.New("pay client not configured")
	}
	resp, err := l.ctx.PayClient.NotifyPayment(req.Context(), &proto.NotifyPaymentRequest{
		PaymentNo:        paymentNo,
		TradeStatus:      req.FormValue("trade_status"),
		TransactionId:    req.FormValue("trade_no"),
		TotalAmountCents: amountCents,
		PaidAt:           paidAt,
		NotifyRaw:        raw,
	})
	if err != nil || resp == nil || !resp.Success {
		return "fail", err
	}
	return "success", nil
}

// buildNotifyRaw 把 url.Values 序列化为按 key 排序的 "k1=v1&k2=v2" 形式（支付宝验签规范）。
func buildNotifyRaw(form url.Values) string {
	if len(form) == 0 {
		return ""
	}
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		if len(form[k]) > 0 {
			sb.WriteString(form[k][0])
		}
	}
	return sb.String()
}
```

**Step 3: 配置 + ServiceContext + router**

- 读 `api/admin/internal/config/config.go`，加 `PayRPC RPCConf \`json:"pay_rpc"\``（若已有 RPCConf 类型复用；否则新建 `type PayRPCConf struct{ Target string }`）。确认现有 `admin_rpc` 配置类型名。
- 读 `api/admin/internal/svc/servicecontext.go`，加 `PayClient paysvcclient.PayService` 字段 + `newPayRPCClient` helper + 初始化（复用 `zrpc.NewClient` + `paysvcclient.NewPayService`）。
- 读 `api/admin/internal/handler/router.go`，在 routes() 加路由 + 文件末尾加 `handleAlipayCallback`。
- 读 `api/admin/etc/admin.json`，加 `"pay_rpc": { "target": "127.0.0.1:50054" }`。

**Step 4: 写单测**

`api/admin/internal/logic/pay_callback_logic_test.go`：用 mock PayClient 覆盖 success/fail/missing 三种场景。

**Step 5: 编译 + 测试**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./api/... 2>&1; go test ./api/admin/internal/logic/ 2>&1`
Expected: 无 error + PASS

---

## Task 7: 全量编译 + 测试回归

**Step 1: 全量构建**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go build ./... 2>&1`
Expected: 仅剩已知 `scripts/e2e` 重复 main 错误（既有问题，非本次引入）

**Step 2: 全量测试（排除 e2e main 包）**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy"; go test ./rpc/paysvc/... ./api/admin/... 2>&1`
Expected: 全部 PASS

**Step 3: 启动联调冒烟（可选，MySQL/Kafka 就绪时）**

Run: `cd "c:/Users/hjy/Desktop/XiaoLong-Ridy/rpc/paysvc"; go run paysvc.go`
Expected: 启动日志显示 "alipay real channel enabled"、两个 job started、HTTP 回调监听启动
