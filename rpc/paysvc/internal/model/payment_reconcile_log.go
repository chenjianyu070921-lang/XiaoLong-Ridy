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
