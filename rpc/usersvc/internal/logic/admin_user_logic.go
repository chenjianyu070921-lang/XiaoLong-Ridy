package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminUserLogic 处理管理后台用户只读查询。
type AdminUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAdminUserLogic 创建管理后台用户查询逻辑实例。
func NewAdminUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserLogic {
	return &AdminUserLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListUsers 按状态分页查询用户，并返回后台所需的基础资料。
func (l *AdminUserLogic) ListUsers(in *userproto.AdminUserListRequest) (*userproto.AdminUserListResponse, error) {
	page, pageSize := in.GetPage(), in.GetPageSize()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := l.svcCtx.Users.ListPage(l.ctx, int(in.GetStatus()), int(page), int(pageSize))
	if err != nil {
		return nil, err
	}
	list := make([]*userproto.AdminUser, 0, len(users))
	for _, user := range users {
		list = append(list, adminUserFromModel(user))
	}
	return &userproto.AdminUserListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetUser 按用户 ID 查询后台所需的乘客用户详情。
// 入参 id 必须大于 0；用户不存在时返回 gRPC NotFound，便于上游统一转换 HTTP 404。
func (l *AdminUserLogic) GetUser(in *userproto.AdminUserDetailRequest) (*userproto.AdminUser, error) {
	if in.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	user, err := l.svcCtx.Users.FindByID(l.ctx, in.GetId())
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	if err != nil {
		return nil, err
	}
	return adminUserFromModel(user), nil
}

// adminUserFromModel 将用户领域模型转换为管理后台专用 RPC 结构。
// usersvc 只负责提供真实领域数据，最终展示脱敏由 adminsvc 按后台权限策略统一处理。
func adminUserFromModel(user *model.User) *userproto.AdminUser {
	if user == nil {
		return nil
	}
	return &userproto.AdminUser{
		Id:             user.ID,
		Phone:          user.Phone,
		Nickname:       user.Nickname,
		AvatarUrl:      user.AvatarURL,
		Gender:         int32(user.Gender),
		RealName:       user.RealName,
		IdCardNo:       user.IDCardNo,
		RegisterSource: user.RegisterSource,
		Status:         int32(user.Status),
		CreatedAt:      user.CreatedAt.Unix(),
		UpdatedAt:      user.UpdatedAt.Unix(),
	}
}
