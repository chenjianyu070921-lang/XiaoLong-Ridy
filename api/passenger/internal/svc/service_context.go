package svc

import (
	"context"

	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// UserClient 定义 passenger API 调用 usersvc 的公开契约。
type UserClient interface {
	SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error)
	LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error)
	RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error)
	Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error)
}

type ServiceContext struct {
	UserClient UserClient
}

func NewServiceContext(userClient UserClient) *ServiceContext {
	return &ServiceContext{UserClient: userClient}
}
