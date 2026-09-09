package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// ExportLogic 负责后台导出任务 HTTP 请求到 adminsvc 的参数转换。
type ExportLogic struct {
	ctx *svc.ServiceContext
}

// NewExportLogic 创建导出任务逻辑对象。
func NewExportLogic(ctx *svc.ServiceContext) *ExportLogic {
	return &ExportLogic{ctx: ctx}
}

// Create 创建导出任务。
func (l *ExportLogic) Create(ctx context.Context, req types.ExportTaskRequest, session *model.AdminSession, ip string) (*adminclient.ExportTaskResponse, error) {
	adminID := int64(0)
	if session != nil {
		adminID = session.AdminID
	}
	return l.ctx.AdminSvc.CreateExportTask(ctx, &adminclient.ExportTaskRequest{
		ExportType: req.ExportType,
		Filters:    req.Filters,
		AdminId:    adminID,
		Ip:         ip,
	})
}

// List 查询导出任务列表。
func (l *ExportLogic) List(ctx context.Context, page, pageSize int, exportType string) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListExportTasks(ctx, &adminclient.ExportTaskListRequest{
		Page:       int32(page),
		PageSize:   int32(pageSize),
		ExportType: exportType,
	})
	if err != nil {
		return nil, err
	}
	return &types.PageResult{List: resp.List, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Detail 查询单个导出任务详情。
// 任务状态、文件路径和失败原因均由 adminsvc 维护，HTTP 层只透传任务编号。
func (l *ExportLogic) Detail(ctx context.Context, taskNo string) (*adminclient.ExportTask, error) {
	return l.ctx.AdminSvc.GetExportTask(ctx, &adminclient.ExportTaskDetailRequest{TaskNo: taskNo})
}
