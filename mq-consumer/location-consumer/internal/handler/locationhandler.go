package handler

import (
	"context"
	"strconv"

	"XiaoLong-Ridy/common/constants"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// LocationHandler 位置消息处理器
type LocationHandler struct {
	rdb *redis.Client
	logx.Logger
}

func NewLocationHandler(rdb *redis.Client) *LocationHandler {
	return &LocationHandler{
		rdb:    rdb,
		Logger: logx.WithContext(context.Background()),
	}
}

// Consume 处理一条位置事件：维护 driver:online 在线司机集合（在线/行程中加入，离线移除），并记录日志
func (h *LocationHandler) Consume(ctx context.Context, id string, values map[string]interface{}) error {
	driverID, _ := strconv.ParseInt(toString(values["driver_id"]), 10, 64)
	onlineStatus, _ := strconv.Atoi(toString(values["online_status"]))
	lat := toFloat64(values["lat"])
	lng := toFloat64(values["lng"])
	city := toString(values["city"])

	member := strconv.FormatInt(driverID, 10)
	if onlineStatus == 0 {
		h.rdb.SRem(ctx, constants.DriverOnlineKey, member)
	} else {
		h.rdb.SAdd(ctx, constants.DriverOnlineKey, member)
	}

	h.Infof("位置事件 id=%s driverId=%d city=%s online=%d lat=%v lng=%v", id, driverID, city, onlineStatus, lat, lng)
	return nil
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return ""
	}
}

func toFloat64(v interface{}) float64 {
	switch x := v.(type) {
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	case float64:
		return x
	default:
		return 0
	}
}
