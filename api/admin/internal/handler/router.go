package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/logic"
	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"

	"google.golang.org/grpc/codes"
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

	r.mux.HandleFunc("/admin/v1/operation-logs", r.authRequired(r.handleOperationLogs))

	r.mux.HandleFunc("/admin/v1/users", r.authRequired(r.handleUsers))
	r.mux.HandleFunc("/admin/v1/users/", r.authRequired(r.handleUserByID))

	r.mux.HandleFunc("/admin/v1/driver-certifications", r.authRequired(r.handleDriverCertifications))
	r.mux.HandleFunc("/admin/v1/driver-certifications/", r.authRequired(r.handleDriverCertificationByID))

	r.mux.HandleFunc("/admin/v1/coupons", r.authRequired(r.handleCoupons))
	r.mux.HandleFunc("/admin/v1/coupons/", r.authRequired(r.handleCouponByID))

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
		s, err := r.ctx.SessionRepository.Get(req.Context(), token)
		if err == nil {
			session = s
		}
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
	session := sessionFromContext(req.Context())
	resp, err := logic.NewAuthLogic(r.ctx).Me(req.Context(), session)
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
	session := sessionFromContext(req.Context())
	items := logic.NewAuthLogic(r.ctx).GetMenus(session)
	writeSuccess(w, map[string]any{"items": items})
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
		writeError(w, http.StatusBadRequest, 40001, "invalid order id")
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

// authRequired 是后台通用鉴权中间件。
// 它从 Authorization 头中读取 token，并从 Redis 中加载会话信息。
func (r *Router) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token := bearerToken(req.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, 40004, "token missing")
			return
		}
		session, err := r.ctx.SessionRepository.Get(req.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, 40004, "token invalid")
			return
		}
		ctx := context.WithValue(req.Context(), sessionContextKey{}, session)
		next(w, req.WithContext(ctx))
	}
}

// writeAuthError 将鉴权相关错误转换为统一响应。
func (r *Router) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, logic.ErrBadRequest):
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case errors.Is(err, logic.ErrUnauthorized), errors.Is(err, repository.ErrAdminNotFound):
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, 40003, "forbidden")
	case errors.Is(err, logic.ErrConflict):
		writeError(w, http.StatusConflict, 40902, "conflict")
	case status.Code(err) == codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.NotFound:
		writeError(w, http.StatusNotFound, 40401, "resource not found")
	case status.Code(err) == codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40003, "forbidden")
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
	case errors.Is(err, repository.ErrUserNotFound),
		errors.Is(err, repository.ErrDriverCertificationNotFound),
		errors.Is(err, repository.ErrOrderNotFound),
		errors.Is(err, repository.ErrCouponNotFound),
		errors.Is(err, repository.ErrOperationLogNotFound):
		writeError(w, http.StatusNotFound, 40401, "resource not found")
	case errors.Is(err, logic.ErrForbidden):
		writeError(w, http.StatusForbidden, 40003, "forbidden")
	case status.Code(err) == codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, 40001, "bad request")
	case status.Code(err) == codes.NotFound:
		writeError(w, http.StatusNotFound, 40401, "resource not found")
	case status.Code(err) == codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40003, "forbidden")
	case status.Code(err) == codes.Unauthenticated:
		writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, 50000, "system error")
	}
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

// clientIP 获取客户端 IP。
func clientIP(req *http.Request) string {
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
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
