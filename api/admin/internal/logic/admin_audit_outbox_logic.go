package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// AdminAuditOutboxLogic 负责通知与审计补偿任务的 HTTP 到 RPC 适配。
type AdminAuditOutboxLogic struct {
	ctx *svc.ServiceContext
}

// NewAdminAuditOutboxLogic 创建补偿任务查询逻辑对象。
func NewAdminAuditOutboxLogic(ctx *svc.ServiceContext) *AdminAuditOutboxLogic {
	return &AdminAuditOutboxLogic{ctx: ctx}
}

// List 查询补偿任务状态、失败原因和重试次数。
func (l *AdminAuditOutboxLogic) List(ctx context.Context, page, pageSize int, statusText, module, action string, targetID int64) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListAdminAuditOutbox(ctx, &adminclient.AdminAuditOutboxListRequest{
		Page: int32(page), PageSize: int32(pageSize), Status: statusText,
		Module: module, Action: action, TargetId: targetID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.AdminAuditOutboxDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.AdminAuditOutboxDTO{
			ID: item.Id, EventNo: item.EventNo, Module: item.Module, Action: item.Action,
			TargetType: item.TargetType, TargetID: item.TargetId, AdminID: item.AdminId,
			Detail: item.Detail, Status: item.Status, RetryCount: item.RetryCount,
			FailureReason: item.FailureReason, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}
