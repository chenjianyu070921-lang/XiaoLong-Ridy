package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// OperationLogLogic 负责后台操作日志查询的 HTTP 到 RPC 适配。
type OperationLogLogic struct {
	ctx *svc.ServiceContext
}

// NewOperationLogLogic 创建操作日志逻辑。
func NewOperationLogLogic(ctx *svc.ServiceContext) *OperationLogLogic {
	return &OperationLogLogic{ctx: ctx}
}

// List 查询后台操作日志列表。
func (l *OperationLogLogic) List(ctx context.Context, req types.OperationLogListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListOperationLogs(ctx, &adminclient.OperationLogListRequest{
		Page: int32(req.Page), PageSize: int32(req.PageSize), AdminId: req.AdminID,
		Module: req.Module, Action: req.Action, TargetType: req.TargetType, TargetId: req.TargetID,
		StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.OperationLogDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.OperationLogDTO{
			ID: item.Id, AdminID: item.AdminId, Module: item.Module, Action: item.Action,
			TargetType: item.TargetType, TargetID: item.TargetId, Detail: item.Detail,
			IP: item.Ip, CreatedAt: item.CreatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}
