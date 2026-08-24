package client

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/model"
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
	coupons := repository.NewMemoryCouponRepository()
	// 本地模式也预置与生产迁移脚本一致的四张新人券，便于完整演示首次登录和预估价扣减。
	now := time.Now()
	for _, coupon := range []*model.Coupon{
		{ID: 9001, Name: "新人首单立减20元", Type: 3, FaceValue: 20, ThresholdAmount: 25, ValidStartAt: now, ValidEndAt: now.Add(90 * 24 * time.Hour), Status: 2, PerUserLimit: 1},
		{ID: 9002, Name: "新人第二单立减8元", Type: 3, FaceValue: 8, ThresholdAmount: 20, ValidStartAt: now, ValidEndAt: now.Add(90 * 24 * time.Hour), Status: 2, PerUserLimit: 1},
		{ID: 9003, Name: "新人第三单立减5元", Type: 3, FaceValue: 5, ThresholdAmount: 20, ValidStartAt: now, ValidEndAt: now.Add(90 * 24 * time.Hour), Status: 2, PerUserLimit: 1},
		{ID: 9004, Name: "夜间出行立减5元", Type: 3, FaceValue: 5, ThresholdAmount: 15, ValidStartAt: now, ValidEndAt: now.Add(90 * 24 * time.Hour), Status: 2, PerUserLimit: 1},
	} {
		coupons.AddCouponForTest(coupon)
	}
	smsService := logic.NewMemorySMSCodeService(onSMSCode)
	tokens := logic.NewTokenManager(signingKey)
	svcCtx := svc.NewServiceContext(config.Config{}, users, addresses, coupons, repository.NewMemoryRiskBlacklistRepository(), smsService, smsService, tokens, nil) // nil = 本地开发环境跳过实名认证
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

// GetProfile 转发个人中心资料查询 RPC。
func (c *LocalClient) GetProfile(ctx context.Context, req *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error) {
	return c.service.GetProfile(ctx, req)
}

// SubmitRealName 转发实名资料提交 RPC。
func (c *LocalClient) SubmitRealName(ctx context.Context, req *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error) {
	return c.service.SubmitRealName(ctx, req)
}

// UpdateProfile 转发个人资料更新 RPC。
func (c *LocalClient) UpdateProfile(ctx context.Context, req *userproto.UpdateProfileRequest) (*userproto.UpdateProfileResponse, error) {
	return c.service.UpdateProfile(ctx, req)
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

// ClaimCoupon 转发领取优惠券 RPC。
func (c *LocalClient) ClaimCoupon(ctx context.Context, req *userproto.ClaimCouponRequest) (*userproto.ClaimCouponResponse, error) {
	return c.service.ClaimCoupon(ctx, req)
}

// ListMyCoupons 转发我的优惠券列表 RPC。
func (c *LocalClient) ListMyCoupons(ctx context.Context, req *userproto.ListMyCouponsRequest) (*userproto.ListMyCouponsResponse, error) {
	return c.service.ListMyCoupons(ctx, req)
}

// LockUserCoupon 转发下单锁券 RPC。
func (c *LocalClient) LockUserCoupon(ctx context.Context, req *userproto.LockUserCouponRequest) (*userproto.LockUserCouponResponse, error) {
	return c.service.LockUserCoupon(ctx, req)
}

// ReleaseUserCoupon 转发释放锁券 RPC。
func (c *LocalClient) ReleaseUserCoupon(ctx context.Context, req *userproto.ReleaseUserCouponRequest) (*userproto.ReleaseUserCouponResponse, error) {
	return c.service.ReleaseUserCoupon(ctx, req)
}
