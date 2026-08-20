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
	registerOrderRoutes(mux, svcCtx)
	registerAddressRoutes(mux, svcCtx)
	registerCouponRoutes(mux, svcCtx)
	registerReviewRoutes(mux, svcCtx)
	return mux
}

// registerAuthRoutes 注册乘客端登录注册相关路由，登录前接口不需要 JWT。
func registerAuthRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 发送短信验证码接口，用于乘客登录前获取一次性验证码。
	mux.HandleFunc("/api/passenger/v1/auth/send-sms-code", handler.SendSMSCodeHandler(svcCtx))
	// 短信验证码登录接口，验证码校验通过后签发乘客 JWT。
	mux.HandleFunc("/api/passenger/v1/auth/login-by-sms", handler.LoginBySMSHandler(svcCtx))
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
}

// registerOrderRoutes 注册乘客订单接口。
func registerOrderRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	// 创建订单接口，完成价格预估、优惠券锁定和订单创建。
	mux.HandleFunc("/api/passenger/v1/orders/create", handler.CreateOrderHandler(svcCtx))
	// 查询订单列表接口，按当前乘客和订单状态分页查询历史订单。
	mux.HandleFunc("/api/passenger/v1/orders/list", handler.ListOrdersHandler(svcCtx))
	// 查询订单详情接口，返回当前乘客指定订单的完整行程信息。
	mux.HandleFunc("/api/passenger/v1/orders/detail", handler.GetOrderHandler(svcCtx))
	mux.HandleFunc("/api/passenger/v1/orders/status", handler.PollOrderStatusHandler(svcCtx))
	// 取消订单接口，由当前乘客发起订单取消操作。
	mux.HandleFunc("/api/passenger/v1/orders/cancel", handler.CancelOrderHandler(svcCtx))
	// 发起支付接口，校验订单归属和待支付状态后调用支付服务。
	mux.HandleFunc("/api/passenger/v1/orders/pay", handler.PayOrderHandler(svcCtx))
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
	// 查询我的优惠券接口，按状态返回当前乘客已领取的优惠券。
	mux.HandleFunc("/api/passenger/v1/coupons/my", handler.ListMyCouponsHandler(svcCtx))
}

func registerReviewRoutes(mux *http.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc("/api/passenger/v1/reviews/submit", handler.SubmitReviewHandler(svcCtx))
}
