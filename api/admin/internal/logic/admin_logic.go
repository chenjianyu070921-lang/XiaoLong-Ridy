package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// AdminLogic 封装管理员管理 HTTP 到 adminsvc 的参数转换。
type AdminLogic struct {
	ctx *svc.ServiceContext
}

// NewAdminLogic 创建管理员管理适配逻辑。
func NewAdminLogic(ctx *svc.ServiceContext) *AdminLogic { return &AdminLogic{ctx: ctx} }

// List 查询管理员列表。
func (l *AdminLogic) List(ctx context.Context, req types.AdminListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListAdmins(ctx, &adminclient.AdminListRequest{Page: int32(req.Page), PageSize: int32(req.PageSize), Keyword: req.Keyword, Role: req.Role, Status: req.Status})
	if err != nil {
		return nil, err
	}
	list := make([]types.AdminDTO, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, toAdminDTO(item))
	}
	return &types.PageResult{List: list, Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())}, nil
}

// Create 新增管理员。
func (l *AdminLogic) Create(ctx context.Context, req types.AdminSaveRequest, session *model.AdminSession, ip string) (types.AdminDTO, error) {
	resp, err := l.ctx.AdminSvc.CreateAdmin(ctx, &adminclient.AdminSaveRequest{Username: req.Username, Password: req.Password, RealName: req.RealName, Role: req.Role, OperatorAdminId: session.AdminID, Ip: ip})
	if err != nil {
		return types.AdminDTO{}, err
	}
	return toAdminDTO(resp), nil
}

// Update 编辑管理员资料。
func (l *AdminLogic) Update(ctx context.Context, id int64, req types.AdminSaveRequest, session *model.AdminSession, ip string) (types.AdminDTO, error) {
	resp, err := l.ctx.AdminSvc.UpdateAdmin(ctx, &adminclient.AdminSaveRequest{Id: id, RealName: req.RealName, Role: req.Role, OperatorAdminId: session.AdminID, Ip: ip})
	if err != nil {
		return types.AdminDTO{}, err
	}
	return toAdminDTO(resp), nil
}

// SetStatus 启用或停用管理员。
func (l *AdminLogic) SetStatus(ctx context.Context, id int64, req types.AdminStatusRequest, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.SetAdminStatus(ctx, &adminclient.AdminStatusRequest{Id: id, Status: req.Status, Reason: req.Reason, OperatorAdminId: session.AdminID, Ip: ip})
	return err
}

// ResetPassword 重置管理员密码。
func (l *AdminLogic) ResetPassword(ctx context.Context, id int64, req types.AdminPasswordResetRequest, session *model.AdminSession, ip string) error {
	if strings.TrimSpace(req.Password) == "" {
		return ErrBadRequest
	}
	_, err := l.ctx.AdminSvc.ResetAdminPassword(ctx, &adminclient.AdminPasswordResetRequest{Id: id, Password: req.Password, OperatorAdminId: session.AdminID, Ip: ip})
	return err
}
