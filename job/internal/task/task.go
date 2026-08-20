package task

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/job/internal/svc"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// driverGeoKey 与 locationsvc 写入 Redis GEO 的 key 保持一致（默认城市）
const driverGeoKey = "driver:geo:default"

// Task 定时任务业务逻辑
type Task struct {
	svcCtx *svc.ServiceContext
}

func NewTask(svcCtx *svc.ServiceContext) *Task {
	return &Task{svcCtx: svcCtx}
}

// CleanExpiredLocation 清理过期司机位置数据（默认保留最近 7 天）
func (t *Task) CleanExpiredLocation(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 7
	}
	ctx := context.Background()
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	// 1. 取出过期记录的 driver_id，用于清理 Redis GEO 成员
	var driverIDs []int64
	if err := t.svcCtx.Db.Table("driver_location").
		Where("created_at < ?", cutoff).
		Pluck("driver_id", &driverIDs).Error; err != nil {
		return fmt.Errorf("查询过期位置失败: %w", err)
	}

	// 2. 从 Redis GEO 移除这些司机（member 为字符串 driver_id，与上报时一致）
	if len(driverIDs) > 0 {
		members := make([]interface{}, 0, len(driverIDs))
		for _, id := range driverIDs {
			members = append(members, fmt.Sprintf("%d", id))
		}
		if err := t.svcCtx.Redis.ZRem(ctx, driverGeoKey, members...).Err(); err != nil {
			logx.Errorf("清理 Redis GEO 失败: %v", err)
		}
	}

	// 3. 删除过期 DB 记录
	res := t.svcCtx.Db.Exec("DELETE FROM driver_location WHERE created_at < ?", cutoff)
	if res.Error != nil {
		return fmt.Errorf("删除过期位置失败: %w", res.Error)
	}

	logx.Infof("清理过期位置数据完成: 删除DB记录 %d 条, 清理GEO成员 %d 个", res.RowsAffected, len(driverIDs))
	return nil
}

// DailyReport 生成每日统计报表并落库
func (t *Task) DailyReport() error {
	ctx := context.Background()
	page, pageSize := int32(1), int32(100)

	// 昨日时间范围（本地时区）
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
	end := start.AddDate(0, 0, 1)

	var total, completed, cancelled, other int64
	for {
		resp, err := t.svcCtx.OrderClient.ListOrders(ctx, &order.ListOrdersRequest{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("拉取订单列表失败: %w", err)
		}
		for _, o := range resp.List {
			ct := time.Unix(o.CreatedAt, 0)
			if ct.Before(start) || ct.After(end) {
				continue
			}
			total++
			switch o.Status {
			case 5: // ORDER_STATUS_COMPLETED
				completed++
			case 6: // ORDER_STATUS_CANCELLED
				cancelled++
			default:
				other++
			}
		}
		if int64(page*pageSize) >= resp.Total {
			break
		}
		page++
	}

	report := &DailyReport{
		ReportDate:      start.Format("2006-01-02"),
		TotalOrders:     total,
		CompletedOrders: completed,
		CancelledOrders: cancelled,
		OtherOrders:     other,
		CreatedAt:       now,
	}
	if err := t.svcCtx.Db.AutoMigrate(&DailyReport{}); err != nil {
		return fmt.Errorf("建日报表失败: %w", err)
	}
	if err := t.svcCtx.Db.Create(report).Error; err != nil {
		return fmt.Errorf("落库日报表失败: %w", err)
	}

	logx.Infof("每日报表(%s): 订单总数=%d 已完成=%d 已取消=%d 其他=%d",
		report.ReportDate, total, completed, cancelled, other)
	return nil
}

// DailyReport 每日统计报表（由 job 自管，落库 daily_report 表）
type DailyReport struct {
	Id              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportDate      string    `gorm:"column:report_date;size:10;index" json:"report_date"`
	TotalOrders     int64     `gorm:"column:total_orders" json:"total_orders"`
	CompletedOrders int64     `gorm:"column:completed_orders" json:"completed_orders"`
	CancelledOrders int64     `gorm:"column:cancelled_orders" json:"cancelled_orders"`
	OtherOrders     int64     `gorm:"column:other_orders" json:"other_orders"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
}

func (DailyReport) TableName() string { return "daily_report" }
