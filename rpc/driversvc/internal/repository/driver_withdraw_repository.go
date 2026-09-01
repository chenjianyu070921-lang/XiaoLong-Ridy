package repository

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
)

// DriverWithdrawRepository 定义司机提现数据访问接口，使 logic 层与具体存储实现解耦。
type DriverWithdrawRepository interface {
	// Create 写入一条提现申请记录。
	Create(ctx context.Context, withdraw *model.DriverWithdraw) error
	// ListByDriver 按司机 ID 分页查询提现记录，按申请时间倒序返回。
	ListByDriver(ctx context.Context, driverID uint64, page, pageSize int32) ([]*model.DriverWithdraw, int64, error)
	// AdminList 按管理后台筛选条件分页查询提现记录，按申请时间倒序返回。
	AdminList(ctx context.Context, filter AdminWithdrawFilter) ([]*model.DriverWithdraw, int64, error)
	// GetByID 按主键查询提现记录。
	GetByID(ctx context.Context, id uint64) (*model.DriverWithdraw, error)
	// Audit 写入提现审核结果：status 为目标状态，paidAt 仅打款成功时写入。
	Audit(ctx context.Context, id uint64, status int32, remark string, paidAt *time.Time) error
}

// AdminWithdrawFilter 管理后台提现列表筛选条件；零值字段表示不过滤。
type AdminWithdrawFilter struct {
	Page     int32
	PageSize int32
	Status   int32
	DriverID uint64
	Keyword  string
}
