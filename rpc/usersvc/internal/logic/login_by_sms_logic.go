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

// LoginBySMSLogic 处理短信验证码登录 RPC 的完整业务流程。
type LoginBySMSLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLoginBySMSLogic 创建短信登录逻辑实例，并绑定请求上下文与服务依赖。
func NewLoginBySMSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginBySMSLogic {
	return &LoginBySMSLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LoginBySMS 校验验证码并完成登录；首次手机号登录时自动创建乘客账号。
func (l *LoginBySMSLogic) LoginBySMS(in *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	if !IsValidPhone(phone) {
		return nil, ErrInvalidPhone
	}

	// 先校验验证码，再进行用户查询或自动注册，避免无效请求污染用户数据。
	ok, err := l.svcCtx.SMSVerifier.Verify(l.ctx, phone, strings.TrimSpace(in.GetCode()))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidSMSCode
	}

	user, isNewUser, err := l.findOrCreateUser(phone)
	if err != nil {
		return nil, err
	}
	// 新账号创建成功后自动发放四张新人券，保证优惠券默认属于账号而不是依赖前端额外点击领取。
	if isNewUser {
		if err := l.issueWelcomeCoupons(user.ID); err != nil {
			return nil, err
		}
	}
	if user.Status == model.UserStatusFrozen {
		return nil, ErrAccountFrozen
	}

	// 登录成功后统一签发 Access Token 和 Refresh Token，供乘客端后续接口鉴权使用。
	token, refreshToken, err := l.svcCtx.Tokens.Issue(user.ID, user.Phone, user.Status)
	if err != nil {
		return nil, err
	}

	return &userproto.LoginBySMSResponse{
		Token:        token,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
		User:         toUserInfo(user),
	}, nil
}

// findOrCreateUser 按手机号查询用户；不存在时按短信登录规则自动注册。
func (l *LoginBySMSLogic) findOrCreateUser(phone string) (*model.User, bool, error) {
	user, err := l.svcCtx.Users.FindByPhone(l.ctx, phone)
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		user = &model.User{
			Phone:          phone,
			Nickname:       MaskPhone(phone),
			RegisterSource: model.RegisterSourcePhone,
			Status:         model.UserStatusNormal,
		}
		if err := l.svcCtx.Users.Create(l.ctx, user); err != nil {
			return nil, false, err
		}
		return user, true, nil
	case err != nil:
		return nil, false, err
	default:
		return user, false, nil
	}
}

// issueWelcomeCoupons 为新账号发放固定的四张新人券。
// 券模板由数据库迁移脚本统一维护；领取接口本身具备幂等限制，避免重复发放。
func (l *LoginBySMSLogic) issueWelcomeCoupons(userID uint64) error {
	if l == nil || l.svcCtx == nil || l.svcCtx.Coupons == nil || userID == 0 {
		return ErrCouponRepositoryNotConfigured
	}
	for _, couponID := range []uint64{9001, 9002, 9003, 9004} {
		if _, err := l.svcCtx.Coupons.Claim(l.ctx, userID, couponID); err != nil {
			return err
		}
	}
	return nil
}

// toUserInfo 将内部用户模型转换为 proto 响应结构，并避免返回明文手机号。
func toUserInfo(user *model.User) *userproto.UserInfo {
	return &userproto.UserInfo{
		UserId:         user.ID,
		Phone:          MaskPhone(user.Phone),
		Nickname:       user.Nickname,
		AvatarUrl:      user.AvatarURL,
		RealNameStatus: realNameStatus(user),
	}
}
