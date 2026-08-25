package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/common/realname"
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
//
// # 处理流程：
// #  1. 参数校验（非空检查）
// #  2. 若已配置腾讯云实名认证，则调用二要素核验接口
// #  3. 核验通过后更新用户表中的实名信息
func (l *SubmitRealNameLogic) SubmitRealName(in *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error) {
	realName := strings.TrimSpace(in.GetRealName())
	idCardNo := strings.TrimSpace(in.GetIdCardNo())
	if in.GetUserId() == 0 || realName == "" || idCardNo == "" {
		return nil, ErrInvalidRealNameInfo
	}

	// 调用腾讯云进行实名认证（若已配置）
	if err := l.verifyRealName(realName, idCardNo); err != nil {
		return nil, err
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

// verifyRealName 调用实名认证服务进行二要素核验，若未配置则跳过。
func (l *SubmitRealNameLogic) verifyRealName(name, idCardNo string) error {
	verifier, err := realNameVerifier(l.svcCtx)
	if err != nil {
		return err
	}

	// 未配置实名认证服务时跳过核验（兼容本地开发环境）
	if verifier == nil {
		l.Logger.Info("未配置实名认证服务，跳过核验")
		return nil
	}

	result, err := verifier.Verify(l.ctx, name, idCardNo)
	if err != nil {
		l.Logger.Errorf("腾讯云实名认证调用失败: %v", err)
		return ErrRealNameVerifyFailed
	}

	// Result="0" 表示姓名和身份证号一致
	if result.Result != "0" {
		l.Logger.Errorf("实名认证未通过: result=%s description=%s", result.Result, result.Description)
		return ErrRealNameVerifyFailed
	}

	l.Logger.Infof("实名认证通过: name=%s", name)
	return nil
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

// realNameVerifier 获取实名认证服务依赖。
func realNameVerifier(svcCtx *svc.ServiceContext) (realname.Verifier, error) {
	if svcCtx == nil {
		return nil, errors.New("service context is nil")
	}
	return svcCtx.RealNameVer, nil
}

// UpdateProfileLogic 处理乘客个人资料更新 RPC。
type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewUpdateProfileLogic 创建个人资料更新逻辑实例。
func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateProfile 更新乘客昵称与头像，空字段表示不修改。
//
// # 处理流程：
// #  1. 参数校验（user_id 非空、昵称 <=20 字且允许特殊符号）
// #  2. 查询现有用户并应用非空字段的更新
// #  3. 持久化后返回更新后的用户基础资料
func (l *UpdateProfileLogic) UpdateProfile(in *userproto.UpdateProfileRequest) (*userproto.UpdateProfileResponse, error) {
	if in.GetUserId() == 0 {
		return nil, userproto.ErrUserNotFound
	}

	if in.GetNickname() != "" {
		nickname := strings.TrimSpace(in.GetNickname())
		// 按字符（rune）计长，支持中文与特殊符号，限制 20 字以内。
		if len([]rune(nickname)) > 20 {
			return nil, userproto.ErrNicknameTooLong
		}
		in.Nickname = nickname
	}

	users, err := userRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	user, err := users.FindByID(l.ctx, in.GetUserId())
	if err != nil {
		return nil, mapUserRepositoryError(err)
	}
	if in.GetNickname() != "" {
		user.Nickname = in.GetNickname()
	}
	if in.GetAvatarUrl() != "" {
		user.AvatarURL = in.GetAvatarUrl()
	}
	if err := users.Update(l.ctx, user); err != nil {
		return nil, mapUserRepositoryError(err)
	}
	return &userproto.UpdateProfileResponse{User: toUserInfo(user)}, nil
}
