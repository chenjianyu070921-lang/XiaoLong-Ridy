package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// ErrDriverNotFound 表示未找到指定的司机记录（含已软删记录）。
var ErrDriverNotFound = errors.New("driver not found")

// DriverRepository 定义司机数据访问接口，使 logic 层与具体存储实现解耦。
type DriverRepository interface {
	// Create 写入一条司机记录。
	Create(ctx context.Context, driver *model.Driver) error
	// GetByID 按主键查询司机（软删记录不可见）。
	GetByID(ctx context.Context, id uint64) (*model.Driver, error)
	// Update 按 ID 增量更新司机字段。
	Update(ctx context.Context, id uint64, updates map[string]interface{}) error
	// Delete 软删除指定司机（设置 deleted_at）。
	Delete(ctx context.Context, driver *model.Driver) error
}
