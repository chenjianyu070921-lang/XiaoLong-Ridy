package client

import (
	"context"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/server"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// LocalClient 是开发环境内存版 usersvc 客户端。
// API 层只依赖 proto 契约；切换为 gRPC/zRPC 时可替换该实现。
type LocalClient struct {
	service *server.UserServer
}

// NewLocalClient 创建本地联调用的 usersvc 客户端。
// onSMSCode 仅供开发环境查看验证码，生产环境应替换为真实短信通道。
func NewLocalClient(signingKey string, onSMSCode func(phone, code string)) *LocalClient {
	users := repository.NewMemoryUserRepository()
	addresses := repository.NewMemoryAddressRepository()
	smsService := logic.NewMemorySMSCodeService(onSMSCode)
	tokens := logic.NewTokenManager(signingKey)
	svcCtx := svc.NewServiceContext(config.Config{}, users, addresses, smsService, smsService, tokens)
	return &LocalClient{
		service: server.NewUserServer(svcCtx),
	}
}

// SendSMSCode 转发发送短信验证码 RPC。
func (c *LocalClient) SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	return c.service.SendSMSCode(ctx, req)
}

// LoginBySMS 转发短信登录 RPC。
func (c *LocalClient) LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	return c.service.LoginBySMS(ctx, req)
}

// RefreshToken 转发刷新令牌 RPC。
func (c *LocalClient) RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	return c.service.RefreshToken(ctx, req)
}

// Logout 转发注销 RPC。
func (c *LocalClient) Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	return c.service.Logout(ctx, req)
}

// CreateAddress 转发新增常用地址 RPC。
func (c *LocalClient) CreateAddress(ctx context.Context, req *userproto.CreateAddressRequest) (*userproto.AddressInfo, error) {
	return c.service.CreateAddress(ctx, req)
}

// ListAddresses 转发查询常用地址列表 RPC。
func (c *LocalClient) ListAddresses(ctx context.Context, req *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error) {
	return c.service.ListAddresses(ctx, req)
}

// UpdateAddress 转发更新常用地址 RPC。
func (c *LocalClient) UpdateAddress(ctx context.Context, req *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error) {
	return c.service.UpdateAddress(ctx, req)
}

// DeleteAddress 转发删除常用地址 RPC。
func (c *LocalClient) DeleteAddress(ctx context.Context, req *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error) {
	return c.service.DeleteAddress(ctx, req)
}
