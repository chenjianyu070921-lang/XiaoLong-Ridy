package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/geo"
	order "XiaoLong-Ridy/rpc/ordersvc/orderclient"

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

type DriverPushDispatchMessage struct {
	Type         string `json:"type"`
	OrderId      int64  `json:"order_id"`
	DriverId     int64  `json:"driver_id"`
	DispatchedAt int64  `json:"dispatched_at"`
	ServerTime   int64  `json:"server_time"`
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
	// 订单终点只查一次：供开启回家模式的司机做顺路判定（终点获取失败不阻断派单）。
	orderToLng, orderToLat := c.orderDestination(ctx, evt.OrderId)
	// 写入每个候选司机的待接单列表（幂等：Set 去重）。
	for _, driverID := range evt.DriverIds {
		if driverID <= 0 {
			continue
		}
		// 回家模式：司机开启后只接收路线顺路的订单，反向单直接不推送。
		if ok, err := c.onRouteForHome(ctx, driverID, orderToLng, orderToLat); err == nil && !ok {
			logx.WithContext(ctx).Infof("dispatch.new filtered by home mode: orderId=%d driverId=%d", evt.OrderId, driverID)
			continue
		}
		key := availableListKey(driverID)
		if err := c.svcCtx.Redis.SAdd(ctx, key, evt.OrderId).Err(); err != nil {
			logx.WithContext(ctx).Errorf("push available order %d to driver %d failed: %v", evt.OrderId, driverID, err)
			continue
		}
		// 待接单列表保留 90s，避免司机端长期堆积过期派单。
		c.svcCtx.Redis.Expire(ctx, key, 90*time.Second)
		pushPayload, err := json.Marshal(DriverPushDispatchMessage{
			Type:         constants.TopicDispatchNew,
			OrderId:      evt.OrderId,
			DriverId:     driverID,
			DispatchedAt: evt.DispatchedAt,
			ServerTime:   time.Now().Unix(),
		})
		if err != nil {
			logx.WithContext(ctx).Errorf("marshal dispatch push for order %d driver %d failed: %v", evt.OrderId, driverID, err)
			continue
		}
		if err := c.svcCtx.Redis.Publish(ctx, fmt.Sprintf(constants.RedisDriverPush, driverID), pushPayload).Err(); err != nil {
			logx.WithContext(ctx).Errorf("publish dispatch push for order %d driver %d failed: %v", evt.OrderId, driverID, err)
			continue
		}
	}
	return nil
}

// orderDestination 查询订单终点坐标；查询失败返回零值（调用方会跳过顺路判定，不阻断派单）。
func (c *OrderConsumer) orderDestination(ctx context.Context, orderID int64) (lng, lat float64) {
	if c.svcCtx == nil || c.svcCtx.OrderClient == nil || orderID <= 0 {
		return 0, 0
	}
	order, err := c.svcCtx.OrderClient.GetOrder(ctx, &order.GetOrderRequest{OrderId: orderID})
	if err != nil || order == nil {
		logx.WithContext(ctx).Errorf("get order destination failed: orderId=%d err=%v", orderID, err)
		return 0, 0
	}
	return order.GetToLongitude(), order.GetToLatitude()
}

// onRouteForHome 判断该订单对司机是否顺路：
// 未开启回家模式或读取设置失败时返回 true（放行），只在明确判定不顺路时返回 false。
func (c *OrderConsumer) onRouteForHome(ctx context.Context, driverID int64, orderToLng, orderToLat float64) (bool, error) {
	if c.svcCtx == nil || c.svcCtx.Redis == nil || driverID <= 0 {
		return true, nil
	}
	homeKey := fmt.Sprintf(constants.RedisDriverHome, driverID)
	home, err := c.svcCtx.Redis.HGetAll(ctx, homeKey).Result()
	if err != nil {
		return true, err
	}
	if len(home) == 0 || home["open"] != "1" {
		return true, nil // 未开启回家模式：全域听单
	}
	homeLng, _ := strconv.ParseFloat(home["lng"], 64)
	homeLat, _ := strconv.ParseFloat(home["lat"], 64)
	ratio, _ := strconv.ParseFloat(home["max_detour_ratio"], 64)
	// 司机当前位置：优先读位置快照，缺失时回退 GEO。
	driverLng, driverLat := 0.0, 0.0
	if pos, posErr := c.svcCtx.Redis.HGetAll(ctx, fmt.Sprintf(constants.RedisDriverPos, driverID)).Result(); posErr == nil {
		driverLng, _ = strconv.ParseFloat(pos["longitude"], 64)
		driverLat, _ = strconv.ParseFloat(pos["latitude"], 64)
	}
	if !geo.Valid(driverLng, driverLat) {
		return true, nil // 拿不到司机位置：不做过滤，避免误杀
	}
	return geo.IsOnRoute(driverLng, driverLat, orderToLng, orderToLat, homeLng, homeLat, ratio), nil
}
