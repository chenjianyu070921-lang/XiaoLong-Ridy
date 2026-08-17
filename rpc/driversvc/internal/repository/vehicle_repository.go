package repository

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// ErrVehicleNotFound 表示未找到指定的车辆记录（含已软删记录）。
var ErrVehicleNotFound = errors.New("vehicle not found")

// DriverVehicleRepository 定义车辆数据访问接口，使 logic 层与具体存储实现解耦。
type DriverVehicleRepository interface {
	// Create 写入一条车辆记录。
	Create(ctx context.Context, vehicle *model.DriverVehicle) error
	// GetByID 按主键查询车辆（软删记录不可见）。
	GetByID(ctx context.Context, id uint64) (*model.DriverVehicle, error)
	// Update 按 ID 增量更新车辆字段。
	Update(ctx context.Context, id uint64, updates map[string]interface{}) error
	// Delete 软删除指定车辆（设置 deleted_at）。
	Delete(ctx context.Context, vehicle *model.DriverVehicle) error
}
