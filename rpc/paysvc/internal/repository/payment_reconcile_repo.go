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
