package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

var ErrUserClientNotConfigured = errors.New("user client not configured")

type AuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendSMSCode 转发发送短信验证码请求。
func (l *AuthLogic) SendSMSCode(req *types.SendSMSCodeRequest) (*types.SendSMSCodeResponse, error) {
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.SendSMSCode(l.ctx, &userproto.SendSMSCodeRequest{
		Phone: strings.TrimSpace(req.Phone),
	})
	if err != nil {
		return nil, err
	}
	return &types.SendSMSCodeResponse{Success: resp.GetSuccess(), ExpireIn: resp.GetExpireIn()}, nil
}

// LoginBySMS 转发短信登录请求，并转换 usersvc 返回的数据结构。
func (l *AuthLogic) LoginBySMS(req *types.LoginBySMSRequest) (*types.LoginBySMSResponse, error) {
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.LoginBySMS(l.ctx, &userproto.LoginBySMSRequest{
		Phone: strings.TrimSpace(req.Phone),
		Code:  strings.TrimSpace(req.Code),
	})
	if err != nil {
		return nil, err
	}

	user := resp.GetUser()
	return &types.LoginBySMSResponse{
		Token:        resp.GetToken(),
		RefreshToken: resp.GetRefreshToken(),
		IsNewUser:    resp.GetIsNewUser(),
		User: types.UserInfo{
			UserID:         user.GetUserId(),
			Phone:          user.GetPhone(),
			Nickname:       user.GetNickname(),
			AvatarURL:      user.GetAvatarUrl(),
			RealNameStatus: user.GetRealNameStatus(),
		},
	}, nil
}

// RefreshToken 转发刷新令牌请求。
func (l *AuthLogic) RefreshToken(req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.RefreshToken(l.ctx, &userproto.RefreshTokenRequest{
		RefreshToken: strings.TrimSpace(req.RefreshToken),
	})
	if err != nil {
		return nil, err
	}
	return &types.RefreshTokenResponse{Token: resp.GetToken(), RefreshToken: resp.GetRefreshToken()}, nil
}

// Logout 转发当前登录态注销请求。
func (l *AuthLogic) Logout(token string) (*types.LogoutResponse, error) {
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Logout(l.ctx, &userproto.LogoutRequest{Token: strings.TrimSpace(token)})
	if err != nil {
		return nil, err
	}
	return &types.LogoutResponse{Success: resp.GetSuccess()}, nil
}

func (l *AuthLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}
