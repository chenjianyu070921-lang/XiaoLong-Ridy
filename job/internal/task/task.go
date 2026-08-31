package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/job/internal/svc"
	dispatch "XiaoLong-Ridy/rpc/dispatchsvc/dispatch"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

	"github.com/redis/go-redis/v9"
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

// dispatchRetryTask 与 ordersvc 入队结构保持一致（JSON tag 相同）。
// 序列化格式变更时需同步 rpc/ordersvc/internal/logic/create_order_logic.go（P1-M4-2）。
type dispatchRetryTask struct {
	OrderId       int64   `json:"order_id"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
	Attempt       int     `json:"attempt"`
}

// refundRetryTask 描述一条待重新投递的订单退款事件。
type refundRetryTask struct {
	OrderId       int64  `json:"order_id"`
	OrderNo       string `json:"order_no"`
	RefundNo      string `json:"refund_no"`
	RefundCents   int64  `json:"refund_cents"`
	OperatorId    int64  `json:"operator_id"`
	OperatorType  string `json:"operator_type"`
	RefundTimeout int64  `json:"refund_timeout"`
	Attempt       int    `json:"attempt"`
}

// RetryRefundEvents 扫描退款事件延迟队列并重新投递 Kafka。
func (t *Task) RetryRefundEvents(max int) error {
	if t.svcCtx.Redis == nil || t.svcCtx.EventProducer == nil {
		return fmt.Errorf("refund event retry dependencies not configured")
	}
	if max <= 0 {
		max = 50
	}
	ctx := context.Background()
	items, err := t.svcCtx.Redis.ZRangeByScore(ctx, constants.RefundRetryQueueKey, &redis.ZRangeBy{
		Min: "-inf", Max: strconv.FormatInt(time.Now().Unix(), 10), Count: int64(max),
	}).Result()
	if err != nil {
		return fmt.Errorf("读取退款事件重试队列失败: %w", err)
	}
	for _, member := range items {
		var item refundRetryTask
		if err := json.Unmarshal([]byte(member), &item); err != nil {
			_ = t.svcCtx.Redis.ZRem(ctx, constants.RefundRetryQueueKey, member).Err()
			continue
		}
		if err := t.svcCtx.EventProducer.Send(constants.TopicOrderRefunded, item.RefundNo, []byte(member)); err == nil {
			_ = t.svcCtx.Redis.ZRem(ctx, constants.RefundRetryQueueKey, member).Err()
			continue
		}
		item.Attempt++
		itemPayload, _ := json.Marshal(item)
		delay := time.Duration(5*int(math.Pow(3, float64(item.Attempt-1)))) * time.Second
		if item.Attempt > constants.MaxRefundRetryAttempt {
			delay = time.Hour
		}
		_ = t.svcCtx.Redis.ZAdd(ctx, constants.RefundRetryQueueKey, redis.Z{
			Score: float64(time.Now().Add(delay).Unix()), Member: string(itemPayload),
		}).Err()
	}
	return nil
}

// RetryPendingDispatches 扫描派单失败延迟重试队列（dispatch:retry:orders），
// 对已到期任务重新调用派单服务：成功移除、失败按指数退避（5s/15s/45s）重排（P1-M4-2）。
// 超过最大尝试次数后保留在队列（score 延后 1h），由告警/人工介入，避免死循环。
// max 限制单轮处理条数，防止积压时一次性拉取过多。
func (t *Task) RetryPendingDispatches(max int) error {
	if t.svcCtx.Redis == nil {
		return fmt.Errorf("redis not configured")
	}
	if max <= 0 {
		max = 50
	}
	ctx := context.Background()
	items, err := t.svcCtx.Redis.ZRangeByScore(ctx, constants.DispatchRetryQueueKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   strconv.FormatInt(time.Now().Unix(), 10),
		Count: int64(max),
	}).Result()
	if err != nil {
		return fmt.Errorf("读取派单重试队列失败: %w", err)
	}

	done := 0
	for _, member := range items {
		var item dispatchRetryTask
		if err := json.Unmarshal([]byte(member), &item); err != nil {
			// 脏数据直接移除并告警，避免阻塞队列。
			_ = t.svcCtx.Redis.ZRem(ctx, constants.DispatchRetryQueueKey, member).Err()
			logx.Errorf("解析派单重试任务失败，已丢弃: %s err=%v", member, err)
			continue
		}
		if _, err := t.svcCtx.DispatchClient.DispatchOrder(ctx, &dispatch.DispatchOrderRequest{
			OrderId:       item.OrderId,
			FromLongitude: item.FromLongitude,
			FromLatitude:  item.FromLatitude,
			CarType:       item.CarType,
			CityCode:      item.CityCode,
		}); err != nil {
			next := &dispatchRetryTask{
				OrderId: item.OrderId, FromLongitude: item.FromLongitude,
				FromLatitude: item.FromLatitude, CarType: item.CarType,
				CityCode: item.CityCode, Attempt: item.Attempt + 1,
			}
			payload, _ := json.Marshal(next)
			if next.Attempt > constants.MaxDispatchRetryAttempt {
				// 重试次数耗尽：延后 1h 再暴露，避免每轮都触发，供告警/人工介入。
				_ = t.svcCtx.Redis.ZAdd(ctx, constants.DispatchRetryQueueKey, redis.Z{
					Score: float64(time.Now().Add(time.Hour).Unix()), Member: string(payload),
				}).Err()
				logx.Errorf("派单补偿重试次数耗尽, order_id=%d attempt=%d, 保留队列待人工介入: %v", item.OrderId, next.Attempt, err)
				continue
			}
			delay := time.Duration(5*int(math.Pow(3, float64(next.Attempt-1)))) * time.Second
			_ = t.svcCtx.Redis.ZAdd(ctx, constants.DispatchRetryQueueKey, redis.Z{
				Score: float64(time.Now().Add(delay).Unix()), Member: string(payload),
			}).Err()
			logx.Errorf("派单补偿重试失败, order_id=%d attempt=%d, 下次 %v 后重试: %v", item.OrderId, next.Attempt, delay, err)
			continue
		}
		_ = t.svcCtx.Redis.ZRem(ctx, constants.DispatchRetryQueueKey, member).Err()
		done++
		logx.Infof("派单补偿成功, order_id=%d", item.OrderId)
	}
	logx.Infof("派单补偿扫描完成: 拉取 %d 条, 成功 %d 条", len(items), done)
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
