package router

import (
	"XiaoLong-Ridy/api/passenger/internal/handler"
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/svc"
)

// NewRouter 创建乘客端 HTTP 路由入口，统一注册当前已实现的认证接口。
func NewRouter(svcCtx *svc.ServiceContext) http.Handler {
	mux := http.NewServeMux()
	registerAuthRoutes(mux, svcCtx)
	registerProfileRoutes(mux, svcCtx)
	registerUploadRoutes(mux, svcCtx)
	registerOrderRoutes(mux, svcCtx)
	registerAddressRoutes(mux, svcCtx)
	registerCouponRoutes(mux, svcCtx)
	registerReviewRoutes(mux, svcCtx)
	registerWalletRoutes(mux, svcCtx)
	return mux
}

// registerWalletRoutes 注册乘客钱包查询、充值和提现接口。
func registerWalletRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc("/api/passenger/v1/wallet", handler.GetWalletHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/wallet/recharge", handler.RechargeWalletHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/wallet/withdraw", handler.WithdrawWalletHandler(svcCtx))
}

// registerUploadRoutes 注册需要登录后申请的文件上传凭证接口。
func registerUploadRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc("/api/passenger/v1/upload/avatar-token", handler.AvatarUploadTokenHandler(svcCtx))
}

// registerAuthRoutes 注册乘客端登录注册相关路由，登录前接口不需要 JWT。
func registerAuthRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 发送短信验证码接口，用于乘客登录前获取一次性验证码。
	mux.HandleFunc("/api/passenger/v1/auth/send-sms-code", handler.SendSMSCodeHandler(svcCtx))
	// 短信验证码登录接口，验证码校验通过后签发乘客 JWT。
	mux.HandleFunc("/api/passenger/v1/auth/login-by-sms", handler.LoginBySMSHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/auth/login-by-password", handler.LoginByPasswordHandler(svcCtx))
	// 刷新登录令牌接口，用 refreshToken 换取新的访问令牌。
	mux.HandleFunc("/api/passenger/v1/auth/refresh-token", handler.RefreshTokenHandler(svcCtx))
	// 退出登录接口，用于注销当前乘客登录态。
	mux.HandleFunc("/api/passenger/v1/auth/logout", handler.LogoutHandler(svcCtx))
}

// registerProfileRoutes 注册乘客个人中心接口。
func registerProfileRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 查询当前乘客个人资料接口，返回手机号、昵称和实名认证状态。
	mux.HandleFunc("/api/passenger/v1/profile/me", handler.GetProfileHandler(svcCtx))
	// 提交实名认证资料接口，保存乘客真实姓名和证件号。
	mux.HandleFunc("/api/passenger/v1/profile/real-name", handler.SubmitRealNameHandler(svcCtx))
	// 更新个人资料接口，支持修改昵称与头像，空字段表示不修改。
	mux.HandleFunc("/api/passenger/v1/profile/update", handler.UpdateProfileHandler(svcCtx))
	// 设置或修改密码接口，仅允许当前已登录乘客调用。
	mux.HandleFunc("/api/passenger/v1/profile/password", handler.SetPasswordHandler(svcCtx))
}

// registerOrderRoutes 注册乘客订单接口。
func registerOrderRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 创建订单接口，完成价格预估、优惠券锁定和订单创建。
	mux.HandleFunc("/api/passenger/v1/orders/create", handler.CreateOrderHandler(svcCtx))
	// 行程费用预估接口，选择起终点和车型后实时返回预估金额，不创建订单。
	mux.HandleFunc("/api/passenger/v1/orders/estimate", handler.EstimateOrderHandler(svcCtx))
	// 查询订单列表接口，按当前乘客和订单状态分页查询历史订单。
	mux.HandleFunc("/api/passenger/v1/orders/list", handler.ListOrdersHandler(svcCtx))
	// 查询订单详情接口，返回当前乘客指定订单的完整行程信息。
	mux.HandleFunc("/api/passenger/v1/orders/detail", handler.GetOrderHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/orders/status", handler.PollOrderStatusHandler(svcCtx))
	// 行程实时追踪接口，聚合订单司机位置和剩余路线，仅允许订单所属乘客查询。
	mux.HandleFunc("/api/passenger/v1/orders/tracking", handler.GetOrderTrackingHandler(svcCtx))
	// 取消订单接口，由当前乘客发起订单取消操作。
	mux.HandleFunc("/api/passenger/v1/orders/cancel", handler.CancelOrderHandler(svcCtx))
	// 发起支付接口，校验订单归属和待支付状态后调用支付服务。
	mux.HandleFunc("/api/passenger/v1/orders/pay", handler.PayOrderHandler(svcCtx))
	// 查询支付状态接口，供前端在支付回调/消费者延迟时主动刷新支付结果。
	mux.HandleFunc("/api/passenger/v1/orders/payment-status", handler.GetPaymentStatusHandler(svcCtx))
	// 查询派单状态接口，供前端在派单推送延迟时主动拉取候选司机记录。
	mux.HandleFunc("/api/passenger/v1/orders/dispatch-status", handler.GetDispatchStatusHandler(svcCtx))
}

// registerAddressRoutes 注册乘客常用地址接口。
func registerAddressRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 新增常用地址接口，保存乘客联系人、地址和坐标信息。
	mux.HandleFunc("/api/passenger/v1/addresses/create", handler.CreateAddressHandler(svcCtx))
	// 查询常用地址列表接口，返回当前乘客保存的地址集合。
	mux.HandleFunc("/api/passenger/v1/addresses/list", handler.ListAddressesHandler(svcCtx))
	// 更新常用地址接口，修改当前乘客指定地址的信息。
	mux.HandleFunc("/api/passenger/v1/addresses/update", handler.UpdateAddressHandler(svcCtx))
	// 删除常用地址接口，删除当前乘客指定地址。
	mux.HandleFunc("/api/passenger/v1/addresses/delete", handler.DeleteAddressHandler(svcCtx))
}

// registerCouponRoutes 注册乘客优惠券接口。
func registerCouponRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 领取优惠券接口，将可领取券模板绑定到当前乘客账户。
	mux.HandleFunc("/api/passenger/v1/coupons/claim", handler.ClaimCouponHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/coupons/welcome-gift", handler.ClaimWelcomeGiftHandler(svcCtx))
	// 查询我的优惠券接口，按状态返回当前乘客已领取的优惠券。
	mux.HandleFunc("/api/passenger/v1/coupons/my", handler.ListMyCouponsHandler(svcCtx))
}

// registerReviewRoutes 注册乘客评价接口。
func registerReviewRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc("/api/passenger/v1/reviews/submit", handler.SubmitReviewHandler(svcCtx))
}
