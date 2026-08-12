package client

import (
	"context"

	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/server"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// LocalClient 是开发环境内存版 usersvc 客户端。
// API 层只依赖 proto 契约；切换为 gRPC/zRPC 时可替换该实现。
type LocalClient struct {
	service *server.UserService
}

// NewLocalClient 创建可用于本地联调的 usersvc 客户端。
// onSMSCode 仅供开发环境查看验证码，生产环境应替换为真实短信通道。
func NewLocalClient(signingKey string, onSMSCode func(phone, code string)) *LocalClient {
	users := repository.NewMemoryUserRepository()
	smsService := logic.NewMemorySMSCodeService(onSMSCode)
	tokens := logic.NewTokenManager(signingKey)
	return &LocalClient{
		service: server.NewUserService(users, smsService, smsService, tokens),
	}
}

func (c *LocalClient) SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	return c.service.SendSMSCode(ctx, req)
}

func (c *LocalClient) LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	return c.service.LoginBySMS(ctx, req)
}

func (c *LocalClient) RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	return c.service.RefreshToken(ctx, req)
}

func (c *LocalClient) Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	return c.service.Logout(ctx, req)
}
