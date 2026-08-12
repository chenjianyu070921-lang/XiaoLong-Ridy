package handler

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// CleanupHandler 定时清理任务
type CleanupHandler struct {
	logx.Logger
}

func NewCleanupHandler() *CleanupHandler {
	return &CleanupHandler{
		Logger: logx.WithContext(context.Background()),
	}
}

// CleanExpiredLocation 清理过期司机位置数据
func (h *CleanupHandler) CleanExpiredLocation() error {
	h.Info("执行定时任务: 清理过期位置数据")
	return nil
}

// SyncOrderStatus 同步异常订单状态
func (h *CleanupHandler) SyncOrderStatus() error {
	h.Info("执行定时任务: 同步异常订单状态")
	return nil
}

// DailyReport 生成每日统计报表
func (h *CleanupHandler) DailyReport() error {
	h.Info("执行定时任务: 生成每日报表")
	return nil
}
