package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// DriverBankCardRepository 定义司机银行卡数据访问接口。
type DriverBankCardRepository interface {
	// Create 写入一张银行卡。
	Create(ctx context.Context, card *model.DriverBankCard) error
	// ListByDriver 按司机 ID 列出全部未软删银行卡，按 id 升序。
	ListByDriver(ctx context.Context, driverID uint64) ([]*model.DriverBankCard, error)
	// CountByDriver 统计司机已绑银行卡数量（不含软删）。
	CountByDriver(ctx context.Context, driverID uint64) (int64, error)
	// GetByID 按主键查询银行卡（不含软删）。
	GetByID(ctx context.Context, id uint64) (*model.DriverBankCard, error)
	// Delete 软删除指定银行卡（设置 deleted_at）。
	Delete(ctx context.Context, id uint64) error
	// UpdateWithdrawPasswordHash 更新司机名下全部银行卡的提现密码哈希（司机级密码冗余同步）。
	UpdateWithdrawPasswordHash(ctx context.Context, driverID uint64, hash string) error
}
