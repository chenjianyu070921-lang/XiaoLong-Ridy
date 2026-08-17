package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// UserLogic 封装管理后台用户查询逻辑。
// HTTP 层只负责参数接入，真正的数据查询下沉到 rpc/adminsvc。
type UserLogic struct {
	ctx *svc.ServiceContext
}

// NewUserLogic 创建用户查询逻辑。
func NewUserLogic(ctx *svc.ServiceContext) *UserLogic {
	return &UserLogic{ctx: ctx}
}

// List 查询用户列表并转换为 API 返回结构。
func (l *UserLogic) List(ctx context.Context, req types.UserListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListUsers(ctx, &adminclient.UserListRequest{
		Page:      int32(req.Page),
		PageSize:  int32(req.PageSize),
		Keyword:   req.Keyword,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.UserDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.UserDTO{
			ID:             item.Id,
			Phone:          item.Phone,
			Nickname:       item.Nickname,
			AvatarURL:      item.AvatarUrl,
			Gender:         item.Gender,
			RealName:       item.RealName,
			IDCardNo:       item.IdCardNo,
			RegisterSource: item.RegisterSource,
			Status:         item.Status,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Detail 查询单个用户详情。
func (l *UserLogic) Detail(ctx context.Context, id int64) (*types.UserDTO, error) {
	resp, err := l.ctx.AdminSvc.GetUser(ctx, &adminclient.UserDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.UserDTO{
		ID:             resp.Id,
		Phone:          resp.Phone,
		Nickname:       resp.Nickname,
		AvatarURL:      resp.AvatarUrl,
		Gender:         resp.Gender,
		RealName:       resp.RealName,
		IDCardNo:       resp.IdCardNo,
		RegisterSource: resp.RegisterSource,
		Status:         resp.Status,
		CreatedAt:      resp.CreatedAt,
		UpdatedAt:      resp.UpdatedAt,
	}, nil
}

// Freeze 冻结用户账号。
func (l *UserLogic) Freeze(ctx context.Context, id int64, reason, remark string, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.FreezeUser(ctx, &adminclient.ChangeUserStatusRequest{
		Id: id, Status: 2, Reason: reason, Remark: remark, AdminId: session.AdminID, Ip: ip,
	})
	return err
}

// Unfreeze 解封用户账号。
func (l *UserLogic) Unfreeze(ctx context.Context, id int64, reason, remark string, session *model.AdminSession, ip string) error {
	_, err := l.ctx.AdminSvc.UnfreezeUser(ctx, &adminclient.ChangeUserStatusRequest{
		Id: id, Status: 1, Reason: reason, Remark: remark, AdminId: session.AdminID, Ip: ip,
	})
	return err
}
