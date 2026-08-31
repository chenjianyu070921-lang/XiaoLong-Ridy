package repository

import (
	"context"
	"errors"
	"fmt"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"gorm.io/gorm"
)

var ErrInsufficientBalance = errors.New("insufficient wallet balance")

// WalletRepository 定义钱包余额和分类流水的原子写入能力。
type WalletRepository interface {
	Change(ctx context.Context, userID uint64, typ, title string, amount float64, orderID uint64) error
	Get(ctx context.Context, userID uint64) (*model.UserWallet, error)
	ListTransactions(ctx context.Context, userID uint64, limit, offset int) ([]*model.UserWalletTransaction, int64, error)
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
		query := tx.Model(&model.UserWallet{}).Where("user_id = ?", userID)
		if amount < 0 {
			query = query.Where("balance >= ?", -amount)
		}
		if result := query.UpdateColumn("balance", gorm.Expr("balance + ?", amount)); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return ErrInsufficientBalance
		}
		return tx.Create(&model.UserWalletTransaction{UserID: userID, Type: typ, Title: title, Amount: amount, OrderID: orderID}).Error
	})
}

// Get 查询用户钱包；不存在时返回零余额钱包并持久化初始记录。
func (r *GormWalletRepository) Get(ctx context.Context, userID uint64) (*model.UserWallet, error) {
	var wallet model.UserWallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		wallet = model.UserWallet{UserID: userID}
		err = r.db.WithContext(ctx).Create(&wallet).Error
	}
	return &wallet, err
}

// ListTransactions 分页查询用户钱包流水，按最新时间倒序返回。
func (r *GormWalletRepository) ListTransactions(ctx context.Context, userID uint64, limit, offset int) ([]*model.UserWalletTransaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.UserWalletTransaction{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	var list []*model.UserWalletTransaction
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}
