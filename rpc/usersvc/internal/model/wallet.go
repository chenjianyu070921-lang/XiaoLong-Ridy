package model

import "time"

// UserWallet 对应用户钱包余额表，金额单位为元。
type UserWallet struct {
	UserID    uint64    `gorm:"column:user_id;primaryKey"`
	Balance   float64   `gorm:"column:balance"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// UserWalletTransaction 对应钱包分类流水表，收入为正、支出为负。
type UserWalletTransaction struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"column:user_id"`
	Type      string    `gorm:"column:type"`
	Amount    float64   `gorm:"column:amount"`
	OrderID   uint64    `gorm:"column:order_id"`
	Title     string    `gorm:"column:title"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (UserWallet) TableName() string            { return "user_wallet" }
func (UserWalletTransaction) TableName() string { return "user_wallet_transaction" }
