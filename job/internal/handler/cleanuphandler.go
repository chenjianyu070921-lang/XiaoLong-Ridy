package handler

import (
	"context"

	"XiaoLong-Ridy/job/internal/svc"
	"XiaoLong-Ridy/job/internal/task"

	"github.com/zeromicro/go-zero/core/logx"
)

// CleanupHandler 定时清理任务
type CleanupHandler struct {
	logx.Logger
	svcCtx *svc.ServiceContext
	task   *task.Task
}

func NewCleanupHandler(svcCtx *svc.ServiceContext) *CleanupHandler {
	return &CleanupHandler{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svcCtx,
		task:   task.NewTask(svcCtx),
	}
}

// CleanExpiredLocation 清理过期司机位置数据
func (h *CleanupHandler) CleanExpiredLocation() error {
	return h.task.CleanExpiredLocation(7)
}

// SyncOrderStatus 同步异常订单状态
func (h *CleanupHandler) SyncOrderStatus() error {
	return h.task.SyncOrderStatus()
}

// DailyReport 生成每日统计报表
func (h *CleanupHandler) DailyReport() error {
	return h.task.DailyReport()
}
