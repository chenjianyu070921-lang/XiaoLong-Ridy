package handler

import (
	"context"
	"encoding/json"

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
}

func NewLocationHandler() *LocationHandler {
	return &LocationHandler{
		Logger: logx.WithContext(context.Background()),
	}
}

// Consume 处理位置消息
func (h *LocationHandler) Consume(ctx context.Context, key, value string) error {
	var msg LocationMessage
	if err := json.Unmarshal([]byte(value), &msg); err != nil {
		h.Errorf("unmarshal location message failed: %v", err)
		return err
	}

	h.Infof("收到司机位置: driverId=%d, lat=%.6f, lng=%.6f, speed=%.1f, orderId=%d",
		msg.DriverID, msg.Lat, msg.Lng, msg.Speed, msg.OrderID)
	return nil
}
