package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// ErrUserRepositoryNotConfigured 表示 usersvc 未注入用户仓储。
	ErrUserRepositoryNotConfigured = errors.New("user repository not configured")
)

// GetProfileLogic 处理个人中心资料查询 RPC。
type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetProfileLogic 创建个人中心资料查询逻辑实例。
func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetProfile 根据用户 ID 查询乘客基础资料，并屏蔽敏感手机号。
func (l *GetProfileLogic) GetProfile(in *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error) {
	if in.GetUserId() == 0 {
		return nil, userproto.ErrUserNotFound
	}
	users, err := userRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	user, err := users.FindByID(l.ctx, in.GetUserId())
	if err != nil {
		return nil, mapUserRepositoryError(err)
	}
	return &userproto.GetProfileResponse{User: toUserInfo(user)}, nil
}

// SubmitRealNameLogic 处理乘客实名资料提交 RPC。
type SubmitRealNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSubmitRealNameLogic 创建实名资料提交逻辑实例。
func NewSubmitRealNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitRealNameLogic {
	return &SubmitRealNameLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// SubmitRealName 校验并保存实名信息，返回更新后的用户基础资料。
func (l *SubmitRealNameLogic) SubmitRealName(in *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error) {
	realName := strings.TrimSpace(in.GetRealName())
	idCardNo := strings.TrimSpace(in.GetIdCardNo())
	if in.GetUserId() == 0 || realName == "" || idCardNo == "" {
		return nil, ErrInvalidRealNameInfo
	}

	users, err := userRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	user, err := users.FindByID(l.ctx, in.GetUserId())
	if err != nil {
		return nil, mapUserRepositoryError(err)
	}
	user.RealName = realName
	user.IDCardNo = idCardNo
	if err := users.Update(l.ctx, user); err != nil {
		return nil, mapUserRepositoryError(err)
	}
	return &userproto.SubmitRealNameResponse{User: toUserInfo(user)}, nil
}

// userRepository 获取 usersvc 用户仓储依赖。
func userRepository(svcCtx *svc.ServiceContext) (repository.UserRepository, error) {
	if svcCtx == nil || svcCtx.Users == nil {
		return nil, ErrUserRepositoryNotConfigured
	}
	return svcCtx.Users, nil
}

// mapUserRepositoryError 将仓储层用户错误转换为 usersvc 对外错误。
func mapUserRepositoryError(err error) error {
	if errors.Is(err, repository.ErrUserNotFound) {
		return userproto.ErrUserNotFound
	}
	return err
}

// realNameStatus 根据实名字段是否完整计算用户认证状态。
func realNameStatus(user *model.User) string {
	if strings.TrimSpace(user.RealName) != "" && strings.TrimSpace(user.IDCardNo) != "" {
		return model.RealNameStatusVerified
	}
	return model.RealNameStatusUnverified
}
