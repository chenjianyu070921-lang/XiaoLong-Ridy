package repository

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// DriverHomeDestinationRepository 定义司机回家目的地设置的数据访问接口。
type DriverHomeDestinationRepository interface {
	// GetByDriver 按司机 ID 查询回家目的地设置；未设置时返回 nil, nil。
	GetByDriver(ctx context.Context, driverID uint64) (*model.DriverHomeDestination, error)
	// Upsert 保存（新增或更新）司机的回家目的地设置，按 driver_id 唯一。
	Upsert(ctx context.Context, setting *model.DriverHomeDestination) error
	// SetModeOpen 仅切换回家顺路模式开关；返回是否命中记录。
	SetModeOpen(ctx context.Context, driverID uint64, open bool) (bool, error)
}
