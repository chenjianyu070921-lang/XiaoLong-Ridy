package server

import (
	"context"

	"XiaoLong-Ridy/rpc/usersvc/internal/logic"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// UserServer 提供 usersvc 的 RPC 方法入口，保持与 goctl 生成 server 层一致的轻量转发职责。
type UserServer struct {
	svcCtx *svc.ServiceContext
	userproto.UnimplementedUserServer
}

// NewUserServer 创建用户服务 RPC 入口实例。
func NewUserServer(svcCtx *svc.ServiceContext) *UserServer {
	return &UserServer{
		svcCtx: svcCtx,
	}
}

// SendSMSCode 转发发送短信验证码请求到对应 logic。
func (s *UserServer) SendSMSCode(ctx context.Context, req *userproto.SendSMSCodeRequest) (*userproto.SendSMSCodeResponse, error) {
	l := logic.NewSendSMSCodeLogic(ctx, s.svcCtx)
	return l.SendSMSCode(req)
}

// LoginBySMS 转发短信登录请求到对应 logic。
func (s *UserServer) LoginBySMS(ctx context.Context, req *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	l := logic.NewLoginBySMSLogic(ctx, s.svcCtx)
	return l.LoginBySMS(req)
}

// LoginByPassword 转发手机号密码登录请求。
func (s *UserServer) LoginByPassword(ctx context.Context, req *userproto.LoginByPasswordRequest) (*userproto.LoginBySMSResponse, error) {
	return logic.NewLoginByPasswordLogic(ctx, s.svcCtx).LoginByPassword(req)
}

// SetPassword 转发已登录乘客的密码设置请求。
func (s *UserServer) SetPassword(ctx context.Context, req *userproto.SetPasswordRequest) (*userproto.SetPasswordResponse, error) {
	l := logic.NewSetPasswordLogic(ctx, s.svcCtx)
	return l.SetPassword(req)
}

// RefreshToken 转发刷新令牌请求到对应 logic。
func (s *UserServer) RefreshToken(ctx context.Context, req *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	l := logic.NewRefreshTokenLogic(ctx, s.svcCtx)
	return l.RefreshToken(req)
}

// Logout 转发退出登录请求到对应 logic。
func (s *UserServer) Logout(ctx context.Context, req *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	l := logic.NewLogoutLogic(ctx, s.svcCtx)
	return l.Logout(req)
}

// GetProfile 转发个人中心资料查询请求到对应 logic。
func (s *UserServer) GetProfile(ctx context.Context, req *userproto.GetProfileRequest) (*userproto.GetProfileResponse, error) {
	l := logic.NewGetProfileLogic(ctx, s.svcCtx)
	return l.GetProfile(req)
}

// SubmitRealName 转发实名资料提交请求到对应 logic。
func (s *UserServer) SubmitRealName(ctx context.Context, req *userproto.SubmitRealNameRequest) (*userproto.SubmitRealNameResponse, error) {
	l := logic.NewSubmitRealNameLogic(ctx, s.svcCtx)
	return l.SubmitRealName(req)
}

// UpdateProfile 转发个人资料更新请求到对应 logic。
func (s *UserServer) UpdateProfile(ctx context.Context, req *userproto.UpdateProfileRequest) (*userproto.UpdateProfileResponse, error) {
	l := logic.NewUpdateProfileLogic(ctx, s.svcCtx)
	return l.UpdateProfile(req)
}

// CreateAddress 转发新增常用地址请求到对应 logic。
func (s *UserServer) CreateAddress(ctx context.Context, req *userproto.CreateAddressRequest) (*userproto.AddressInfo, error) {
	l := logic.NewCreateAddressLogic(ctx, s.svcCtx)
	return l.CreateAddress(req)
}

// ListAddresses 转发查询常用地址列表请求到对应 logic。
func (s *UserServer) ListAddresses(ctx context.Context, req *userproto.ListAddressesRequest) (*userproto.ListAddressesResponse, error) {
	l := logic.NewListAddressesLogic(ctx, s.svcCtx)
	return l.ListAddresses(req)
}

// UpdateAddress 转发更新常用地址请求到对应 logic。
func (s *UserServer) UpdateAddress(ctx context.Context, req *userproto.UpdateAddressRequest) (*userproto.AddressInfo, error) {
	l := logic.NewUpdateAddressLogic(ctx, s.svcCtx)
	return l.UpdateAddress(req)
}

// DeleteAddress 转发删除常用地址请求到对应 logic。
func (s *UserServer) DeleteAddress(ctx context.Context, req *userproto.DeleteAddressRequest) (*userproto.DeleteAddressResponse, error) {
	l := logic.NewDeleteAddressLogic(ctx, s.svcCtx)
	return l.DeleteAddress(req)
}

// ClaimCoupon 转发领取优惠券请求到对应 logic。
func (s *UserServer) ClaimCoupon(ctx context.Context, req *userproto.ClaimCouponRequest) (*userproto.ClaimCouponResponse, error) {
	l := logic.NewClaimCouponLogic(ctx, s.svcCtx)
	return l.ClaimCoupon(req)
}

// ListMyCoupons 转发查询我的优惠券请求到对应 logic。
func (s *UserServer) ListMyCoupons(ctx context.Context, req *userproto.ListMyCouponsRequest) (*userproto.ListMyCouponsResponse, error) {
	l := logic.NewListMyCouponsLogic(ctx, s.svcCtx)
	return l.ListMyCoupons(req)
}

// LockUserCoupon 转发下单锁券请求到对应 logic。
func (s *UserServer) LockUserCoupon(ctx context.Context, req *userproto.LockUserCouponRequest) (*userproto.LockUserCouponResponse, error) {
	l := logic.NewLockUserCouponLogic(ctx, s.svcCtx)
	return l.LockUserCoupon(req)
}

// ReleaseUserCoupon 转发释放锁券请求到对应 logic。
func (s *UserServer) ReleaseUserCoupon(ctx context.Context, req *userproto.ReleaseUserCouponRequest) (*userproto.ReleaseUserCouponResponse, error) {
	l := logic.NewReleaseUserCouponLogic(ctx, s.svcCtx)
	return l.ReleaseUserCoupon(req)
}

// AdminListUsers 转发管理后台用户列表查询请求。
func (s *UserServer) AdminListUserCoupons(ctx context.Context, req *userproto.AdminListUserCouponsRequest) (*userproto.AdminListUserCouponsResponse, error) {
	l := logic.NewAdminListUserCouponsLogic(ctx, s.svcCtx)
	return l.AdminListUserCoupons(req)
}

func (s *UserServer) AdminListUsers(ctx context.Context, req *userproto.AdminUserListRequest) (*userproto.AdminUserListResponse, error) {
	l := logic.NewAdminUserLogic(ctx, s.svcCtx)
	return l.ListUsers(req)
}

// AdminGetUser 转发管理后台用户详情查询请求。
func (s *UserServer) AdminGetUser(ctx context.Context, req *userproto.AdminUserDetailRequest) (*userproto.AdminUser, error) {
	l := logic.NewAdminUserLogic(ctx, s.svcCtx)
	return l.GetUser(req)
}

// GetWallet 查询当前用户钱包余额与流水。
func (s *UserServer) GetWallet(ctx context.Context, req *userproto.GetWalletRequest) (*userproto.GetWalletResponse, error) {
	return logic.NewWalletLogic(ctx, s.svcCtx).GetWallet(req)
}

// RechargeWallet 执行钱包充值。
func (s *UserServer) RechargeWallet(ctx context.Context, req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error) {
	return logic.NewWalletLogic(ctx, s.svcCtx).RechargeWallet(req)
}

// WithdrawWallet 执行钱包提现。
func (s *UserServer) WithdrawWallet(ctx context.Context, req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error) {
	return logic.NewWalletLogic(ctx, s.svcCtx).WithdrawWallet(req)
}
