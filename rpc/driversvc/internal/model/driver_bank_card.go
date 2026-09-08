package model

import "time"

// DriverBankCard 对应 driver_bank_card 表：司机绑定的银行卡。
// 合规红线：card_no 存储 AES-GCM 密文，任何出口只允许返回脱敏后的前3后3视图；
// 全程不采集、不存储银行卡取款密码。
type DriverBankCard struct {
	Id                   uint64     `gorm:"primaryKey;column:id" json:"id"`
	DriverId             uint64     `gorm:"column:driver_id;index:idx_driver" json:"driverId"`
	BankName             string     `gorm:"column:bank_name;size:50" json:"bankName"`
	CardNo               string     `gorm:"column:card_no;size:255" json:"cardNo"`                 // AES-GCM 密文
	CardNoHash           string     `gorm:"column:card_no_hash;size:64" json:"cardNoHash"`         // SHA-256 查重
	HolderName           string     `gorm:"column:holder_name;size:50" json:"holderName"`          // 持卡人姓名（与司机实名一致）
	HolderIdCard         string     `gorm:"column:holder_id_card;size:30" json:"holderIdCard"`     // 持卡人身份证（与司机实名一致）
	ReservedPhone        string     `gorm:"column:reserved_phone;size:20" json:"reservedPhone"`    // 银行预留手机号
	WithdrawPasswordHash string     `gorm:"column:withdraw_password_hash;size:255" json:"-"`       // 司机级平台提现密码 bcrypt（冗余）
	Status               int8       `gorm:"column:status;default:1" json:"status"`                 // 1正常 0禁用
	CreatedAt            time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt            time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt            *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

// TableName 返回对应的数据库表名。
func (DriverBankCard) TableName() string {
	return "driver_bank_card"
}
