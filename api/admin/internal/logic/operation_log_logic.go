package logic

import (
	"context"
	"fmt"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// OperationLogLogic 封装后台操作日志查询逻辑。
// 它主要服务于审计和问题排查场景，负责将数据库记录转换成接口返回结构。
type OperationLogLogic struct {
	ctx *svc.ServiceContext
}

// NewOperationLogLogic 创建操作日志逻辑对象。
func NewOperationLogLogic(ctx *svc.ServiceContext) *OperationLogLogic {
	return &OperationLogLogic{ctx: ctx}
}

// List 查询操作日志列表。
func (l *OperationLogLogic) List(ctx context.Context, req types.OperationLogListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.OperationLogRepository.List(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]types.OperationLogDTO, 0, len(list))
	for _, item := range list {
		items = append(items, toOperationLogDTO(item))
	}
	return &types.PageResult{
		List:     items,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// toOperationLogDTO 将日志模型转换为对外 DTO。
func toOperationLogDTO(item model.OperationLog) types.OperationLogDTO {
	return types.OperationLogDTO{
		ID:         item.ID,
		AdminID:    item.AdminID,
		Module:     item.Module,
		Action:     item.Action,
		TargetType: item.TargetType,
		TargetID:   item.TargetID,
		Detail:     item.Detail,
		IP:         item.IP,
		CreatedAt:  repository.FormatTime(item.CreatedAt),
	}
}

// formatOperationLogFilter 用于未来调试时输出条件摘要。
// 当前 P0 没有直接暴露给接口，仅保留给排错时使用。
func formatOperationLogFilter(req types.OperationLogListRequest) string {
	return fmt.Sprintf("admin_id=%d module=%s action=%s target_type=%s target_id=%d", req.AdminID, req.Module, req.Action, req.TargetType, req.TargetID)
}
