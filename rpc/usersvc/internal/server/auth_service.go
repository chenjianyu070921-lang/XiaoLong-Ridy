package server

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

var authPhonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

type UserService struct {
	users       repository.UserRepository
	smsSender   logic.SMSCodeSender
	smsVerifier logic.SMSCodeVerifier
	tokens      *logic.TokenManager
}

// NewUserService 创建用户认证服务。
func NewUserService(
	users repository.UserRepository,
	smsSender logic.SMSCodeSender,
	smsVerifier logic.SMSCodeVerifier,
	tokens *logic.TokenManager,
) *UserService {
	return &UserService{
		users:       users,
		smsSender:   smsSender,
		smsVerifier: smsVerifier,
		tokens:      tokens,
	}
}

// SendSMSCode 校验手机号后向短信服务发送验证码。
func (s *UserService) SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	phone := strings.TrimSpace(req.GetPhone())
	if !authPhonePattern.MatchString(phone) {
		return nil, logic.ErrInvalidPhone
	}
	expireIn, err := s.smsSender.Send(ctx, phone)
	if err != nil {
		return nil, err
	}
	return &userproto.SendSMSCodeResponse{Success: true, ExpireIn: expireIn}, nil
}

// LoginBySMS 校验验证码并完成登录；首次手机号登录时自动创建用户。
func (s *UserService) LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	phone := strings.TrimSpace(req.GetPhone())
	if !authPhonePattern.MatchString(phone) {
		return nil, logic.ErrInvalidPhone
	}

	ok, err := s.smsVerifier.Verify(ctx, phone, strings.TrimSpace(req.GetCode()))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, logic.ErrInvalidSMSCode
	}

	// 验证码通过后按手机号查询账号；不存在则按文档约定自动注册。
	user, err := s.users.FindByPhone(ctx, phone)
	isNewUser := false
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		isNewUser = true
		user = &model.User{
			Phone:          phone,
			Nickname:       maskPhone(phone),
			RegisterSource: model.RegisterSourcePhone,
			Status:         model.UserStatusNormal,
		}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	case user.Status == model.UserStatusFrozen:
		return nil, logic.ErrAccountFrozen
	}

	// 无论新老用户，均签发一组新的 Access Token 和 Refresh Token。
	token, refreshToken, err := s.tokens.Issue(user.ID, user.Phone, user.Status)
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

// RefreshToken 根据 Refresh Token 轮换新的登录令牌。
func (s *UserService) RefreshToken(_ context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	// Refresh 内部会删除旧 Refresh Token，实现一次性使用和令牌轮换。
	token, refreshToken, err := s.tokens.Refresh(strings.TrimSpace(req.GetRefreshToken()))
	if err != nil {
		return nil, err
	}
	return &userproto.RefreshTokenResponse{Token: token, RefreshToken: refreshToken}, nil
}

// Logout 废弃当前 Access Token。
func (s *UserService) Logout(_ context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	// 当前开发实现将 Access Token 写入注销集合；生产环境应替换为 Redis 黑名单。
	if err := s.tokens.Revoke(strings.TrimSpace(req.GetToken())); err != nil {
		return nil, err
	}
	return &userproto.LogoutResponse{Success: true}, nil
}

func toUserInfo(user *model.User) *userproto.UserInfo {
	return &userproto.UserInfo{
		UserId:         user.ID,
		Phone:          maskPhone(user.Phone),
		Nickname:       user.Nickname,
		AvatarUrl:      user.AvatarURL,
		RealNameStatus: "UNVERIFIED",
	}
}

func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
