package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

// DispatchNewEvent 与 dispatchsvc 发布的 dispatch.new 事件字段保持一致（统一 snake_case）。
type DispatchNewEvent struct {
	OrderId       int64   `json:"order_id"`
	DriverIds     []int64 `json:"driver_ids"`
	FromLongitude float64 `json:"from_longitude"`
	FromLatitude  float64 `json:"from_latitude"`
	CarType       int32   `json:"car_type"`
	CityCode      string  `json:"city_code"`
	DispatchedAt  int64   `json:"dispatched_at"`
}

// availableListKey 返回派给指定司机的待接单集合 key。
// 约定：司机端 B 的 /orders/available 接口读取该 key 返回派给自己的单。
// key 统一由 common/constants.RedisDriverAvailable 定义，避免多端硬编码不一致（P2-M4-9）。
func availableListKey(driverID int64) string {
	return fmt.Sprintf(constants.RedisDriverAvailable, driverID)
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
