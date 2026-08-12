package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// UserLogic 封装管理后台用户查询业务。
// P0 阶段只提供只读能力，避免后台绕过用户服务直接修改用户关键状态。
type UserLogic struct {
	ctx *svc.ServiceContext
}

// NewUserLogic 创建用户业务逻辑对象。
func NewUserLogic(ctx *svc.ServiceContext) *UserLogic {
	return &UserLogic{ctx: ctx}
}

// List 查询用户列表，并将数据库模型转换为接口 DTO。
func (l *UserLogic) List(ctx context.Context, req types.UserListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.UserRepository.List(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]types.UserDTO, 0, len(list))
	for _, item := range list {
		items = append(items, toUserDTO(item))
	}
	return &types.PageResult{
		List:     items,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// Detail 查询单个用户详情。
func (l *UserLogic) Detail(ctx context.Context, id int64) (*types.UserDTO, error) {
	user, err := l.ctx.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*user)
	return &dto, nil
}

// toUserDTO 将用户数据库模型转换为接口返回结构。
func toUserDTO(user model.User) types.UserDTO {
	return types.UserDTO{
		ID:             user.ID,
		Phone:          user.Phone,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		Gender:         user.Gender,
		RealName:       user.RealName,
		IDCardNo:       user.IDCardNo,
		RegisterSource: user.RegisterSource,
		Status:         user.Status,
		CreatedAt:      repository.FormatTime(user.CreatedAt),
		UpdatedAt:      repository.FormatTime(user.UpdatedAt),
	}
}
