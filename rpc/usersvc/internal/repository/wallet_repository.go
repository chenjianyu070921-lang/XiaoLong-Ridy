package repository

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"gorm.io/gorm"
)

// WalletRepository 定义钱包余额和分类流水的原子写入能力。
type WalletRepository interface {
	Change(ctx context.Context, userID uint64, typ, title string, amount float64, orderID uint64) error
}

// GormWalletRepository 使用 MySQL 事务更新余额并写入流水，避免余额与明细不一致。
type GormWalletRepository struct{ db *gorm.DB }

func NewGormWalletRepository(db *gorm.DB) *GormWalletRepository { return &GormWalletRepository{db: db} }

func (r *GormWalletRepository) Change(ctx context.Context, userID uint64, typ, title string, amount float64, orderID uint64) error {
	if userID == 0 || amount == 0 {
		return fmt.Errorf("invalid wallet change")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wallet := model.UserWallet{UserID: userID}
		if err := tx.Where("user_id = ?", userID).FirstOrCreate(&wallet).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.UserWallet{}).Where("user_id = ?", userID).UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserWalletTransaction{UserID: userID, Type: typ, Title: title, Amount: amount, OrderID: orderID}).Error
	})
}
