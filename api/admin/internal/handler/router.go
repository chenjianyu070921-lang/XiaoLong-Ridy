package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/logic"
	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminpb "XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Router 是管理后台 API 的 HTTP 路由入口。
// 它负责注册所有 admin 网关路由，并把请求分发到对应的业务逻辑。
type Router struct {
	ctx *svc.ServiceContext
	mux *http.ServeMux
}

// NewRouter 创建管理后台路由对象。
func NewRouter(ctx *svc.ServiceContext) http.Handler {
	r := &Router{
		ctx: ctx,
		mux: http.NewServeMux(),
	}
	r.routes()
	return r
}

// ServeHTTP 让 Router 实现 http.Handler 接口。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// routes 统一注册后台需要的所有路由。
func (r *Router) routes() {
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		writeSuccess(w, map[string]any{
			"service": "XiaoLong-Ridy admin api",
			"health":  "/healthz",
			"auth": map[string]string{
				"register": "POST /admin/v1/auth/register",
				"login":    "POST /admin/v1/auth/login",
				"me":       "GET /admin/v1/auth/me",
				"menus":    "GET /admin/v1/menus",
				"logout":   "POST /admin/v1/auth/logout",
			},
		})
	})

	r.mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"code": 0, "message": "ok"})
	})

	r.mux.HandleFunc("/admin/v1/auth/register", r.handleRegister)
	r.mux.HandleFunc("/admin/v1/auth/login", r.handleLogin)
	r.mux.HandleFunc("/admin/v1/auth/logout", r.authRequired(r.handleLogout))
	r.mux.HandleFunc("/admin/v1/auth/me", r.authRequired(r.handleMe))
	r.mux.HandleFunc("/admin/v1/menus", r.authRequired(r.handleMenus))
	r.mux.HandleFunc("/admin/v1/admins", r.authRequired(r.handleAdmins))
	r.mux.HandleFunc("/admin/v1/admins/", r.authRequired(r.handleAdminByID))

	r.mux.HandleFunc("/admin/v1/operation-logs", r.authRequired(r.handleOperationLogs))

	r.mux.HandleFunc("/admin/v1/users", r.authRequired(r.handleUsers))
	r.mux.HandleFunc("/admin/v1/users/", r.authRequired(r.handleUserByID))

	r.mux.HandleFunc("/admin/v1/driver-certifications", r.authRequired(r.handleDriverCertifications))
	r.mux.HandleFunc("/admin/v1/driver-certifications/", r.authRequired(r.handleDriverCertificationByID))

	r.mux.HandleFunc("/admin/v1/coupons", r.authRequired(r.handleCoupons))
	r.mux.HandleFunc("/admin/v1/coupon-issue-tasks", r.authRequired(r.handleCouponIssueTasks))
	r.mux.HandleFunc("/admin/v1/coupons/", r.authRequired(r.handleCouponByID))

	r.mux.HandleFunc("/admin/v1/price-rules", r.authRequired(r.handlePriceRules))
	r.mux.HandleFunc("/admin/v1/price-rules/", r.authRequired(r.handlePriceRuleByID))

	r.mux.HandleFunc("/admin/v1/promotion-activities", r.authRequired(r.handlePromotionActivities))
	r.mux.HandleFunc("/admin/v1/promotion-activities/", r.authRequired(r.handlePromotionActivityByID))

	r.mux.HandleFunc("/admin/v1/statistics/overview", r.authRequired(r.handleStatisticsOverview))
	r.mux.HandleFunc("/admin/v1/statistics/orders", r.authRequired(r.handleStatisticsOrders))
	r.mux.HandleFunc("/admin/v1/statistics/drivers", r.authRequired(r.handleStatisticsDrivers))
	r.mux.HandleFunc("/admin/v1/statistics/revenue", r.authRequired(r.handleStatisticsRevenue))
	r.mux.HandleFunc("/admin/v1/statistics/coupons", r.authRequired(r.handleStatisticsCoupons))

	r.mux.HandleFunc("/admin/v1/export-tasks", r.authRequired(r.handleExportTasks))
	r.mux.HandleFunc("/admin/v1/export-tasks/", r.authRequired(r.handleExportTaskByNo))
	r.mux.HandleFunc("/admin/v1/work-orders", r.authRequired(r.handleWorkOrders))
	r.mux.HandleFunc("/admin/v1/work-orders/batch-actions", r.authRequired(r.handleWorkOrderBatchActions))
	r.mux.HandleFunc("/admin/v1/work-orders/", r.authRequired(r.handleWorkOrderByID))

	r.mux.HandleFunc("/admin/v1/blacklist", r.authRequired(r.handleBlacklists))
	r.mux.HandleFunc("/admin/v1/blacklist/", r.authRequired(r.handleBlacklistByID))
	r.mux.HandleFunc("/admin/v1/risk/hit-records", r.authRequired(r.handleRiskHitRecords))
	r.mux.HandleFunc("/admin/v1/risk/hit-records/actions", r.authRequired(r.handleRiskHitRecordActions))

	r.mux.HandleFunc("/admin/v1/orders", r.authRequired(r.handleOrders))
	r.mux.HandleFunc("/admin/v1/orders/abnormal", r.authRequired(r.handleAbnormalOrders))
	r.mux.HandleFunc("/admin/v1/orders/", r.authRequired(r.handleOrderByID))
}

// handleRegister 处理后台管理员注册。
// 首个管理员可免登录注册，后续管理员注册需要超级管理员 token。
func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.RegisterRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	var session *model.AdminSession
	if token := bearerToken(req.Header.Get("Authorization")); token != "" {
		s, err := logic.NewAuthLogic(r.ctx).ValidateSession(req.Context(), token)
		if err != nil {
			r.writeAuthError(w, err)
			return
		}
		session = s
		// 注册请求不经过 authRequired，但携带 token 时仍需向 adminsvc 透传，供服务端复核操作者身份。
		req = req.WithContext(metadata.AppendToOutgoingContext(req.Context(), "x-admin-token", token))
	}

	resp, err := logic.NewAuthLogic(r.ctx).Register(req.Context(), &body, session)
	if err != nil {
		r.writeAuthError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleLogin 处理后台管理员登录。
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.LoginRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	resp, err := logic.NewAuthLogic(r.ctx).Login(req.Context(), &body)
	if err != nil {
		r.writeAuthError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleLogout 处理退出登录。
func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	token := bearerToken(req.Header.Get("Authorization"))
	if err := logic.NewAuthLogic(r.ctx).Logout(req.Context(), token); err != nil {
		r.writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"message": "ok"})
}

// handleMe 返回当前登录管理员信息。
func (r *Router) handleMe(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	token := bearerToken(req.Header.Get("Authorization"))
	resp, err := logic.NewAuthLogic(r.ctx).Me(req.Context(), token)
	if err != nil {
		r.writeAuthError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleMenus 返回当前角色对应菜单权限。
func (r *Router) handleMenus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	token := bearerToken(req.Header.Get("Authorization"))
	items, err := logic.NewAuthLogic(r.ctx).GetMenus(req.Context(), token)
	if err != nil {
		r.writeAuthError(w, err)
		return
	}
	writeSuccess(w, map[string]any{"items": items})
}

// handleAdmins 处理超级管理员列表和新增接口。
func (r *Router) handleAdmins(w http.ResponseWriter, req *http.Request) {
	adminLogic := logic.NewAdminLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		query := req.URL.Query()
		resp, err := adminLogic.List(req.Context(), types.AdminListRequest{
			Page: intQuery(req, "page", 1), PageSize: intQuery(req, "page_size", 20),
			Keyword: query.Get("keyword"), Role: int32Query(req, "role", 0), Status: int32Query(req, "status", 0),
		})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.AdminSaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		admin, err := adminLogic.Create(req.Context(), body, sessionFromContext(req.Context()), clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, admin)
	default:
		writeMethodNotAllowed(w)
	}
}

// handleAdminByID 处理管理员编辑、启停和重置密码。
func (r *Router) handleAdminByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/admins/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid admin id")
		return
	}
	adminLogic := logic.NewAdminLogic(r.ctx)
	session := sessionFromContext(req.Context())
	switch action {
	case "":
		if req.Method != http.MethodPut {
			writeMethodNotAllowed(w)
			return
		}
		var body types.AdminSaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		admin, err := adminLogic.Update(req.Context(), id, body, session, clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, admin)
	case "status":
		if req.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}
		var body types.AdminStatusRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := adminLogic.SetStatus(req.Context(), id, body, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	case "reset-password":
		if req.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}
		var body types.AdminPasswordResetRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := adminLogic.ResetPassword(req.Context(), id, body, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	default:
		http.NotFound(w, req)
	}
}

// handleOperationLogs 查询后台操作日志。
func (r *Router) handleOperationLogs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewOperationLogLogic(r.ctx).List(req.Context(), types.OperationLogListRequest{
		Page:       intQuery(req, "page", 1),
		PageSize:   intQuery(req, "page_size", 20),
		AdminID:    int64Query(req, "admin_id", 0),
		Module:     query.Get("module"),
		Action:     query.Get("action"),
		TargetType: query.Get("target_type"),
		TargetID:   int64Query(req, "target_id", 0),
		StartTime:  query.Get("start_time"),
		EndTime:    query.Get("end_time"),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleUsers 返回用户列表。
func (r *Router) handleUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewUserLogic(r.ctx).List(req.Context(), types.UserListRequest{
		Page:      intQuery(req, "page", 1),
		PageSize:  intQuery(req, "page_size", 20),
		Keyword:   query.Get("keyword"),
		Status:    int32Query(req, "status", 0),
		StartTime: query.Get("start_time"),
		EndTime:   query.Get("end_time"),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleUserByID 返回用户详情。
func (r *Router) handleUserByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/users/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid user id")
		return
	}
	userLogic := logic.NewUserLogic(r.ctx)
	if req.Method == http.MethodGet && action == "orders" {
		resp, err := userLogic.OrderHistory(req.Context(), id, intQuery(req, "page", 1), intQuery(req, "page_size", 20), int32Query(req, "status", 0))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if req.Method == http.MethodGet && action == "coupons" {
		resp, err := userLogic.CouponHistory(req.Context(), id, intQuery(req, "page", 1), intQuery(req, "page_size", 20), int32Query(req, "status", 0))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if action == "" {
		if req.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		resp, err := userLogic.Detail(req.Context(), id)
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body struct {
		Reason string `json:"reason"`
		Remark string `json:"remark"`
	}
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	session := sessionFromContext(req.Context())
	switch action {
	case "freeze":
		if err := userLogic.Freeze(req.Context(), id, body.Reason, body.Remark, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	case "unfreeze":
		if err := userLogic.Unfreeze(req.Context(), id, body.Reason, body.Remark, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	default:
		http.NotFound(w, req)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleDriverCertifications 返回司机审核列表。
func (r *Router) handleDriverCertifications(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewDriverLogic(r.ctx).ListCertifications(req.Context(), types.DriverCertificationListRequest{
		Page:        intQuery(req, "page", 1),
		PageSize:    intQuery(req, "page_size", 20),
		Keyword:     query.Get("keyword"),
		AuditStatus: int32Query(req, "audit_status", 0),
		StartTime:   query.Get("start_time"),
		EndTime:     query.Get("end_time"),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleDriverCertificationByID 处理司机审核详情、通过和驳回。
func (r *Router) handleDriverCertificationByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/driver-certifications/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid certification id")
		return
	}
	driverLogic := logic.NewDriverLogic(r.ctx)
	if action == "" {
		if req.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		resp, err := driverLogic.CertificationDetail(req.Context(), id)
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}

	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.AuditRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	session := sessionFromContext(req.Context())
	switch action {
	case "approve":
		if err := driverLogic.ApproveCertification(req.Context(), id, body, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	case "reject":
		if err := driverLogic.RejectCertification(req.Context(), id, body, session, clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	default:
		http.NotFound(w, req)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleOrders 返回订单列表。
func (r *Router) handleOrders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewOrderLogic(r.ctx).List(req.Context(), types.OrderListRequest{
		Page:      intQuery(req, "page", 1),
		PageSize:  intQuery(req, "page_size", 20),
		Keyword:   query.Get("keyword"),
		Status:    int32Query(req, "status", 0),
		UserID:    int64Query(req, "user_id", 0),
		DriverID:  int64Query(req, "driver_id", 0),
		StartTime: query.Get("start_time"),
		EndTime:   query.Get("end_time"),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleAbnormalOrders 查询异常订单列表。
// 该接口用于运营后台定位取消、支付失败、退款和派单异常订单，不执行任何订单写操作。
func (r *Router) handleAbnormalOrders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewOrderLogic(r.ctx).ListAbnormal(req.Context(), types.AbnormalOrderListRequest{
		Page:         intQuery(req, "page", 1),
		PageSize:     intQuery(req, "page_size", 20),
		Keyword:      query.Get("keyword"),
		AbnormalType: query.Get("abnormal_type"),
		UserID:       int64Query(req, "user_id", 0),
		DriverID:     int64Query(req, "driver_id", 0),
		StartTime:    query.Get("start_time"),
		EndTime:      query.Get("end_time"),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleOrderByID 处理订单详情和后台取消订单。
func (r *Router) handleOrderByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/orders/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid orderclient id")
		return
	}
	orderLogic := logic.NewOrderLogic(r.ctx)
	if action == "" {
		if req.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		resp, err := orderLogic.Detail(req.Context(), id)
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if action != "cancel" {
		http.NotFound(w, req)
		return
	}
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.OrderCancelRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if err := orderLogic.Cancel(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleCoupons 处理优惠券模板列表和新增。
// GET 用于列表查询，POST 用于新增模板；新增动作会写入后台操作日志。
func (r *Router) handleCoupons(w http.ResponseWriter, req *http.Request) {
	couponLogic := logic.NewCouponLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		query := req.URL.Query()
		resp, err := couponLogic.List(req.Context(), types.CouponListRequest{
			Page:      intQuery(req, "page", 1),
			PageSize:  intQuery(req, "page_size", 20),
			Keyword:   query.Get("keyword"),
			Type:      int32Query(req, "type", 0),
			Status:    int32Query(req, "status", 0),
			StartTime: query.Get("start_time"),
			EndTime:   query.Get("end_time"),
		})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.CouponSaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := couponLogic.Create(req.Context(), body, sessionFromContext(req.Context()), clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	default:
		writeMethodNotAllowed(w)
	}
}

// handleCouponByID 处理优惠券模板编辑。
// 当前只开放 PUT /admin/v1/coupons/{id}，状态启停和发券任务后续独立扩展。
func (r *Router) handleCouponByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/coupons/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid coupon id")
		return
	}
	couponLogic := logic.NewCouponLogic(r.ctx)
	if action == "disable" {
		if req.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}
		if err := couponLogic.Disable(req.Context(), id, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
		return
	}
	if action == "issue" {
		if req.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}
		var body types.CouponIssueRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := couponLogic.Issue(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if action != "" {
		http.NotFound(w, req)
		return
	}
	if req.Method != http.MethodPut {
		writeMethodNotAllowed(w)
		return
	}
	var body types.CouponSaveRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if err := couponLogic.Update(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleCouponIssueTasks 查询优惠券发放任务。
func (r *Router) handleCouponIssueTasks(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewCouponLogic(r.ctx).ListIssueTasks(req.Context(), types.CouponListRequest{
		Page:      intQuery(req, "page", 1),
		PageSize:  intQuery(req, "page_size", 20),
		StartTime: query.Get("start_time"),
		EndTime:   query.Get("end_time"),
	}, int64Query(req, "coupon_id", 0), query.Get("status"))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handlePriceRules 处理计价规则列表和新增。
func (r *Router) handlePriceRules(w http.ResponseWriter, req *http.Request) {
	priceRuleLogic := logic.NewPriceRuleLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		query := req.URL.Query()
		resp, err := priceRuleLogic.List(req.Context(), types.PriceRuleListRequest{
			Page:     intQuery(req, "page", 1),
			PageSize: intQuery(req, "page_size", 20),
			Keyword:  query.Get("keyword"),
			CityCode: query.Get("city_code"),
			CarType:  int32Query(req, "car_type", 0),
			Status:   int32Query(req, "status", 0),
		})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.PriceRuleSaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := priceRuleLogic.Create(req.Context(), body, sessionFromContext(req.Context()), clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	default:
		writeMethodNotAllowed(w)
	}
}

// handlePriceRuleByID 处理计价规则详情、编辑、启用和停用。
func (r *Router) handlePriceRuleByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/price-rules/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid price rule id")
		return
	}
	priceRuleLogic := logic.NewPriceRuleLogic(r.ctx)
	if action == "" {
		switch req.Method {
		case http.MethodGet:
			resp, err := priceRuleLogic.Detail(req.Context(), id)
			if err != nil {
				r.writeBizError(w, err)
				return
			}
			writeSuccess(w, resp)
		case http.MethodPut:
			var body types.PriceRuleSaveRequest
			if err := decodeJSON(req, &body); err != nil {
				writeError(w, http.StatusBadRequest, 40001, "invalid request body")
				return
			}
			if err := priceRuleLogic.Update(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
				r.writeBizError(w, err)
				return
			}
			writeSuccess(w, types.CommonResponse{Message: "ok"})
		default:
			writeMethodNotAllowed(w)
		}
		return
	}
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	switch action {
	case "enable":
		if err := priceRuleLogic.Enable(req.Context(), id, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	case "disable":
		if err := priceRuleLogic.Disable(req.Context(), id, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	default:
		http.NotFound(w, req)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handlePromotionActivities 处理活动配置列表和新增。
func (r *Router) handlePromotionActivities(w http.ResponseWriter, req *http.Request) {
	marketingLogic := logic.NewMarketingLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		query := req.URL.Query()
		resp, err := marketingLogic.ListActivities(req.Context(), types.PromotionActivityListRequest{
			Page:      intQuery(req, "page", 1),
			PageSize:  intQuery(req, "page_size", 20),
			Keyword:   query.Get("keyword"),
			Type:      int32Query(req, "type", 0),
			Status:    int32Query(req, "status", 0),
			StartTime: query.Get("start_time"),
			EndTime:   query.Get("end_time"),
		})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.PromotionActivitySaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := marketingLogic.CreateActivity(req.Context(), body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	default:
		writeMethodNotAllowed(w)
	}
}

// handlePromotionActivityByID 处理活动配置编辑、发布和回滚。
func (r *Router) handlePromotionActivityByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/promotion-activities/")
	if !ok {
		writeError(w, http.StatusBadRequest, 40001, "invalid activity id")
		return
	}
	marketingLogic := logic.NewMarketingLogic(r.ctx)
	if action == "" {
		if req.Method != http.MethodPut {
			writeMethodNotAllowed(w)
			return
		}
		var body types.PromotionActivitySaveRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := marketingLogic.UpdateActivity(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
		return
	}
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.PromotionActivityActionRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	switch action {
	case "publish":
		if err := marketingLogic.PublishActivity(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	case "rollback":
		if err := marketingLogic.RollbackActivity(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
	default:
		http.NotFound(w, req)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleStatisticsOverview 查询运营总览统计。
func (r *Router) handleStatisticsOverview(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Overview(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleStatisticsOrders 查询订单统计。
func (r *Router) handleStatisticsOrders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Orders(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleStatisticsDrivers 查询司机经营统计。
func (r *Router) handleStatisticsDrivers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Drivers(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleStatisticsRevenue 查询财务收入统计。
func (r *Router) handleStatisticsRevenue(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Revenue(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleStatisticsCoupons 查询优惠券统计。
func (r *Router) handleStatisticsCoupons(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Coupons(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleExportTasks 处理导出任务创建和列表查询。
func (r *Router) handleExportTasks(w http.ResponseWriter, req *http.Request) {
	exportLogic := logic.NewExportLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		resp, err := exportLogic.List(req.Context(), intQuery(req, "page", 1), intQuery(req, "page_size", 20), req.URL.Query().Get("export_type"))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.ExportTaskRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := exportLogic.Create(req.Context(), body, sessionFromContext(req.Context()), clientIP(req))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	default:
		writeMethodNotAllowed(w)
	}
}

// handleExportTaskByNo 查询单个导出任务详情。
func (r *Router) handleExportTaskByNo(w http.ResponseWriter, req *http.Request) {
	taskNo := strings.Trim(strings.TrimPrefix(req.URL.Path, "/admin/v1/export-tasks/"), "/")
	if strings.HasSuffix(taskNo, "/download") {
		taskNo = strings.TrimSuffix(taskNo, "/download")
		r.handleExportDownload(w, req, taskNo)
		return
	}
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if taskNo == "" || strings.Contains(taskNo, "/") {
		writeError(w, http.StatusBadRequest, 40001, "invalid task no")
		return
	}
	resp, err := logic.NewExportLogic(r.ctx).Detail(req.Context(), taskNo)
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleExportDownload 校验下载授权后以流式响应返回 CSV，绝不向客户端暴露服务端文件路径。
func (r *Router) handleExportDownload(w http.ResponseWriter, req *http.Request, taskNo string) {
	if req.Method != http.MethodGet || taskNo == "" || strings.Contains(taskNo, "/") {
		writeError(w, http.StatusBadRequest, 40001, "invalid task no")
		return
	}
	session := sessionFromContext(req.Context())
	if session == nil {
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
		return
	}
	stream, err := adminpb.NewAdminServiceClient(r.ctx.AdminRPCClient.Conn()).DownloadExport(req.Context(), &adminpb.ExportDownloadRequest{TaskNo: taskNo, AdminId: session.AdminID, AdminRole: session.Role})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+taskNo+`.csv"`)
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}
		if _, err = w.Write(chunk.GetContent()); err != nil {
			return
		}
	}
}

// handleWorkOrders 处理后台工单创建和列表查询。
func (r *Router) handleWorkOrders(w http.ResponseWriter, req *http.Request) {
	session := sessionFromContext(req.Context())
	switch req.Method {
	case http.MethodGet:
		resp, err := r.ctx.AdminSvc.ListWorkOrders(req.Context(), &adminclient.WorkOrderListRequest{Page: int32(intQuery(req, "page", 1)), PageSize: int32(intQuery(req, "page_size", 20)), Status: int32Query(req, "status", 0), AssigneeId: int64Query(req, "assignee_id", 0), WorkOrderType: int32Query(req, "work_order_type", 0)})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.PageResult{List: resp.GetList(), Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())})
	case http.MethodPost:
		var body types.WorkOrderRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := r.ctx.AdminSvc.CreateWorkOrder(req.Context(), &adminclient.WorkOrderRequest{WorkOrderType: body.WorkOrderType, SourceType: body.SourceType, SourceId: body.SourceID, OrderId: body.OrderID, UserId: body.UserID, DriverId: body.DriverID, Title: body.Title, Content: body.Content, Priority: body.Priority, AdminId: session.AdminID, Ip: clientIP(req)})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	default:
		writeMethodNotAllowed(w)
	}
}

// handleWorkOrderBatchActions 处理后台工单批量分配、跟进、仲裁、结案和重开。
func (r *Router) handleWorkOrderBatchActions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.WorkOrderBatchActionRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	session := sessionFromContext(req.Context())
	resp, err := r.ctx.AdminSvc.BatchActWorkOrders(req.Context(), &adminclient.WorkOrderBatchActionRequest{
		Ids:               body.IDs,
		Action:            body.Action,
		AssigneeId:        body.AssigneeID,
		Content:           body.Content,
		ArbitrationResult: body.ArbitrationResult,
		AdminId:           session.AdminID,
		Ip:                clientIP(req),
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleWorkOrderByID 处理工单详情、流转与证据索引接口。
func (r *Router) handleWorkOrderByID(w http.ResponseWriter, req *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(req.URL.Path, "/admin/v1/work-orders/"), "/")
	parts := strings.Split(rest, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, 40001, "invalid work order id")
		return
	}
	session := sessionFromContext(req.Context())
	if len(parts) == 1 && req.Method == http.MethodGet {
		resp, err := r.ctx.AdminSvc.GetWorkOrder(req.Context(), &adminclient.WorkOrderDetailRequest{Id: id})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if len(parts) == 2 && parts[1] == "actions" && req.Method == http.MethodPost {
		var body types.WorkOrderActionRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		resp, err := r.ctx.AdminSvc.ActWorkOrder(req.Context(), &adminclient.WorkOrderActionRequest{Id: id, Action: body.Action, AssigneeId: body.AssigneeID, Content: body.Content, ArbitrationResult: body.ArbitrationResult, Version: body.Version, AdminId: session.AdminID, Ip: clientIP(req)})
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
		return
	}
	if len(parts) == 2 && parts[1] == "evidence" {
		if req.Method == http.MethodGet {
			resp, err := r.ctx.AdminSvc.ListWorkOrderEvidence(req.Context(), &adminclient.WorkOrderEvidenceListRequest{WorkOrderId: id, Page: int32(intQuery(req, "page", 1)), PageSize: int32(intQuery(req, "page_size", 20))})
			if err != nil {
				r.writeBizError(w, err)
				return
			}
			writeSuccess(w, types.PageResult{List: resp.GetList(), Total: resp.GetTotal(), Page: int(resp.GetPage()), PageSize: int(resp.GetPageSize())})
			return
		}
		if req.Method == http.MethodPost {
			var body types.WorkOrderEvidenceRequest
			if err := decodeJSON(req, &body); err != nil {
				writeError(w, http.StatusBadRequest, 40001, "invalid request body")
				return
			}
			resp, err := r.ctx.AdminSvc.AddWorkOrderEvidence(req.Context(), &adminclient.WorkOrderEvidenceRequest{WorkOrderId: id, EvidenceType: body.EvidenceType, EvidenceUrl: body.EvidenceURL, Content: body.Content, AdminId: session.AdminID, Ip: clientIP(req)})
			if err != nil {
				r.writeBizError(w, err)
				return
			}
			writeSuccess(w, resp)
			return
		}
	}
	http.NotFound(w, req)
}

// handleBlacklists 处理风控黑名单列表和新增。
func (r *Router) handleBlacklists(w http.ResponseWriter, req *http.Request) {
	riskLogic := logic.NewRiskLogic(r.ctx)
	switch req.Method {
	case http.MethodGet:
		query := req.URL.Query()
		resp, err := riskLogic.ListBlacklists(req.Context(), intQuery(req, "page", 1), intQuery(req, "page_size", 20),
			query.Get("target_type"), int64Query(req, "target_id", 0), int32Query(req, "status", 0))
		if err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, resp)
	case http.MethodPost:
		var body types.BlacklistRequest
		if err := decodeJSON(req, &body); err != nil {
			writeError(w, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := riskLogic.AddBlacklist(req.Context(), body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
			r.writeBizError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	default:
		writeMethodNotAllowed(w)
	}
}

// handleBlacklistByID 处理风控黑名单解除。
func (r *Router) handleBlacklistByID(w http.ResponseWriter, req *http.Request) {
	id, action, ok := idAndActionFromPath(req.URL.Path, "/admin/v1/blacklist/")
	if !ok || action != "release" {
		http.NotFound(w, req)
		return
	}
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.BlacklistRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if err := logic.NewRiskLogic(r.ctx).ReleaseBlacklist(req.Context(), id, body, sessionFromContext(req.Context()), clientIP(req)); err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleRiskHitRecords 查询风控命中记录。
func (r *Router) handleRiskHitRecords(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := req.URL.Query()
	resp, err := logic.NewRiskLogic(r.ctx).ListRiskHitRecords(req.Context(), intQuery(req, "page", 1), intQuery(req, "page_size", 20),
		query.Get("target_type"), int64Query(req, "target_id", 0), query.Get("scene"), int32Query(req, "risk_level", 0))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// handleRiskHitRecordActions 处理风控命中记录复核、拉黑和转工单。
func (r *Router) handleRiskHitRecordActions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.RiskHitActionRequest
	if err := decodeJSON(req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	resp, err := logic.NewRiskLogic(r.ctx).HandleRiskHitRecords(req.Context(), body, sessionFromContext(req.Context()), clientIP(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}

// authRequired 是后台通用鉴权中间件。
// 它只从 Authorization 头中读取 token，并调用 adminsvc 校验会话，Redis 访问统一收敛在 adminsvc。
func (r *Router) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := validateNumericQueryParams(req); err != nil {
			writeError(w, http.StatusBadRequest, 40001, err.Error())
			return
		}
		token := bearerToken(req.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, 40004, "token missing")
			return
		}
		session, err := logic.NewAuthLogic(r.ctx).ValidateSession(req.Context(), token)
		if err != nil {
			r.writeAuthError(w, err)
			return
		}
		// 将已校验 token 作为 gRPC metadata 向下游透传，adminsvc 会再次校验并执行服务端 RBAC。
		ctx := metadata.AppendToOutgoingContext(req.Context(), "x-admin-token", token)
		ctx = context.WithValue(ctx, sessionContextKey{}, session)
		next(w, req.WithContext(ctx))
	}
}

// writeAuthError 将鉴权相关错误转换为统一响应。
func (r *Router) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrBadRequest):
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case errors.Is(err, logic.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, 40003, "forbidden")
	case errors.Is(err, logic.ErrConflict):
		writeError(w, http.StatusConflict, 40902, "conflict")
	case status.Code(err) == codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.FailedPrecondition:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.NotFound:
		writeError(w, http.StatusNotFound, 40401, "resource not found")
	case status.Code(err) == codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40003, permissionDeniedMessage(err))
	case status.Code(err) == codes.Unauthenticated:
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, 50000, "system error")
	}
}

// writeBizError 将业务错误转换为统一响应。
func (r *Router) writeBizError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrBadRequest):
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, 40003, "forbidden")
	case status.Code(err) == codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.FailedPrecondition:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.NotFound:
		writeError(w, http.StatusNotFound, 40401, "resource not found")
	case status.Code(err) == codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40003, permissionDeniedMessage(err))
	case status.Code(err) == codes.Unauthenticated:
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, 50000, "system error")
	}
}

// permissionDeniedMessage 提取 gRPC PermissionDenied 的服务端原因。
// 返回空或非 gRPC 错误时使用通用文案，保证 403 响应始终给出可读原因。
func permissionDeniedMessage(err error) string {
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.PermissionDenied && s.Message() != "" {
			return s.Message()
		}
	}
	return "forbidden"
}

// sessionContextKey 用于在 request context 中保存管理员会话。
type sessionContextKey struct{}

// sessionFromContext 读取已认证的管理员会话。
func sessionFromContext(ctx context.Context) *model.AdminSession {
	if v := ctx.Value(sessionContextKey{}); v != nil {
		if s, ok := v.(*model.AdminSession); ok {
			return s
		}
	}
	return nil
}

// bearerToken 从 Authorization 头解析 Bearer token。
func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

// intQuery 读取整数查询参数。
func intQuery(req *http.Request, key string, defaultValue int) int {
	val := req.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// validateNumericQueryParams 统一校验后台接口中的数字查询参数。
// 缺省值仍由各业务 handler 处理；一旦调用方显式传入非法数字或越界分页参数，直接返回参数错误，
// 避免把拼写错误静默转换成无筛选查询。
func validateNumericQueryParams(req *http.Request) error {
	query := req.URL.Query()
	for _, key := range []string{"page", "page_size", "role", "status", "admin_id", "target_id", "user_id", "driver_id", "audit_status", "type", "coupon_id", "car_type", "assignee_id", "work_order_type", "risk_level"} {
		value, exists := query[key]
		if !exists || len(value) == 0 || strings.TrimSpace(value[0]) == "" {
			continue
		}
		// 发券任务状态使用 pending/processing/success/failed 字符串枚举，
		// 不能套用其他列表接口的数字状态校验。
		if key == "status" && strings.HasPrefix(req.URL.Path, "/admin/v1/coupon-issue-tasks") {
			switch value[0] {
			case "pending", "processing", "success", "failed":
				continue
			}
		}
		parsed, err := strconv.ParseInt(strings.TrimSpace(value[0]), 10, 64)
		if err != nil {
			return errors.New("invalid numeric query parameter: " + key)
		}
		switch key {
		case "page":
			if parsed < 1 {
				return errors.New("page must be greater than zero")
			}
		case "page_size":
			if parsed < 1 || parsed > 100 {
				return errors.New("page_size must be between 1 and 100")
			}
		default:
			if parsed < 0 {
				return errors.New(key + " must not be negative")
			}
		}
	}
	return nil
}

// int32Query 读取 int32 查询参数。
func int32Query(req *http.Request, key string, defaultValue int32) int32 {
	return int32(intQuery(req, key, int(defaultValue)))
}

// int64Query 读取 int64 查询参数。
func int64Query(req *http.Request, key string, defaultValue int64) int64 {
	val := req.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// statisticsRequestFromQuery 读取统计接口通用查询参数。
func statisticsRequestFromQuery(req *http.Request) types.StatisticsRequest {
	query := req.URL.Query()
	return types.StatisticsRequest{
		StartTime: query.Get("start_time"),
		EndTime:   query.Get("end_time"),
		CityCode:  query.Get("city_code"),
	}
}

// idFromPath 从路径中提取 ID。
func idFromPath(path, prefix string) (int64, bool) {
	id, action, ok := idAndActionFromPath(path, prefix)
	return id, ok && action == ""
}

// idAndActionFromPath 从路径中提取 ID 和可选动作。
// 例如 /admin/v1/driver-certifications/12/approve 会返回 12 和 approve。
func idAndActionFromPath(path, prefix string) (int64, string, bool) {
	rest := strings.TrimPrefix(path, prefix)
	if rest == path || rest == "" {
		return 0, "", false
	}
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", false
	}
	if len(parts) == 1 {
		return id, "", true
	}
	if len(parts) == 2 {
		return id, parts[1], true
	}
	return 0, "", false
}

// clientIP 获取客户端直连地址。
// 当前服务未配置可信反向代理列表，因此不信任客户端可直接伪造的转发请求头，
// 避免后台操作审计记录被伪造。部署层如接入可信代理，应在代理层完成地址标准化。
func clientIP(req *http.Request) string {
	host := req.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		return host[:idx]
	}
	return host
}

// decodeJSON 解析 JSON 请求体。
func decodeJSON(req *http.Request, v any) error {
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// writeMethodNotAllowed 输出 405。
func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, 40001, "method not allowed")
}

// writeSuccess 输出统一成功响应。
func writeSuccess(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

// writeError 输出统一错误响应。
func writeError(w http.ResponseWriter, status int, code int, message string) {
	writeJSON(w, status, map[string]any{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}

// writeJSON 写入 JSON 响应体。
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
