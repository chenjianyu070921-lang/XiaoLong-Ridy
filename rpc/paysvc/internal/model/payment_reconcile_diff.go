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
