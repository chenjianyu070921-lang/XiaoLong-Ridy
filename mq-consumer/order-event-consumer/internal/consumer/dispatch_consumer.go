package consumer

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// DispatchNewEvent 与 dispatchsvc 发布的 dispatch.new 事件字段保持一致。
type DispatchNewEvent struct {
	OrderId       int64   `json:"orderId"`
	DriverIds     []int64 `json:"driverIds"`
	FromLongitude float64 `json:"fromLongitude"`
	FromLatitude  float64 `json:"fromLatitude"`
	CarType       int32   `json:"carType"`
	CityCode      string  `json:"cityCode"`
	DispatchedAt  int64   `json:"dispatchedAt"`
}

// availableListKey 司机待接单列表（Redis Set，成员为 orderId）。
// 约定：司机端 B 的 /orders/available 接口读取该 key 返回派给自己的单。
func availableListKey(driverID int64) string {
	return "driver:available:" + strconv.FormatInt(driverID, 10)
}

// handleDispatchNew 将派单结果推送到目标司机的待接单列表，供司机端实时拉取。
func (c *OrderConsumer) handleDispatchNew(ctx context.Context, payload []byte) error {
	var evt DispatchNewEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return err
	}
	if len(evt.DriverIds) == 0 {
		return nil
	}
	// 写入每个候选司机的待接单列表（幂等：Set 去重）。
	for _, driverID := range evt.DriverIds {
		if driverID <= 0 {
			continue
		}
		key := availableListKey(driverID)
		if err := c.svcCtx.Redis.SAdd(ctx, key, evt.OrderId).Err(); err != nil {
			logx.WithContext(ctx).Errorf("push available order %d to driver %d failed: %v", evt.OrderId, driverID, err)
			continue
		}
		// 待接单列表保留 90s，避免司机端长期堆积过期派单。
		c.svcCtx.Redis.Expire(ctx, key, 90*time.Second)
	}
	return nil
}
