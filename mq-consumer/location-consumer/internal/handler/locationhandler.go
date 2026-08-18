package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"XiaoLong-Ridy/common/constants"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// LocationMessage 司机位置消息结构
type LocationMessage struct {
	DriverID  int64   `json:"driver_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Speed     float64 `json:"speed"`
	Direction float64 `json:"direction"`
	Timestamp int64   `json:"timestamp"`
	OrderID   int64   `json:"order_id"`
}

// LocationHandler 位置消息处理器
type LocationHandler struct {
	logx.Logger
	rdb *redis.Client
}

func NewLocationHandler(rdb *redis.Client) *LocationHandler {
	return &LocationHandler{
		Logger: logx.WithContext(context.Background()),
		rdb:    rdb,
	}
}

// Consume 处理位置消息并写入 Redis GEO，供派单引擎检索。
func (h *LocationHandler) Consume(ctx context.Context, _ string, value []byte) error {
	var msg LocationMessage
	if err := json.Unmarshal(value, &msg); err != nil {
		h.Errorf("unmarshal location message failed: %v", err)
		return err
	}

	geoKey := fmt.Sprintf(constants.RedisDriverGeo, "default")
	_, err := h.rdb.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      strconv.FormatInt(msg.DriverID, 10),
		Longitude: msg.Lng,
		Latitude:  msg.Lat,
	}).Result()
	if err != nil {
		h.Errorf("write driver geo failed: %v", err)
		return err
	}
	_, _ = h.rdb.SAdd(ctx, constants.RedisDriverOnline, strconv.FormatInt(msg.DriverID, 10)).Result()

	h.Infof("收到司机位置并写入 GEO: driverId=%d, lat=%.6f, lng=%.6f, speed=%.1f, orderId=%d",
		msg.DriverID, msg.Lat, msg.Lng, msg.Speed, msg.OrderID)
	return nil
}
