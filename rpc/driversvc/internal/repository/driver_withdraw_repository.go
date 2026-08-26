package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// DriverWithdrawRepository 定义司机提现数据访问接口，使 logic 层与具体存储实现解耦。
type DriverWithdrawRepository interface {
	// Create 写入一条提现申请记录。
	Create(ctx context.Context, withdraw *model.DriverWithdraw) error
	// ListByDriver 按司机 ID 分页查询提现记录，按申请时间倒序返回。
	ListByDriver(ctx context.Context, driverID uint64, page, pageSize int32) ([]*model.DriverWithdraw, int64, error)
}
