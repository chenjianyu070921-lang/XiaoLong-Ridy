package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// AcceptOrderHandler POST /api/driver/v1/orders/accept
// 当前登录司机接单。需携带有效 JWT。
func AcceptOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.AcceptOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "orderId 不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).AcceptOrder(int64(claims.AccountID), req.OrderID)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// StartTripHandler POST /api/driver/v1/orders/start-trip
// 当前登录司机开始行程。需携带有效 JWT。
func StartTripHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.StartTripRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "orderId 不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).StartTrip(int64(claims.AccountID), req.OrderID)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// FinishTripHandler POST /api/driver/v1/orders/finish-trip
// 当前登录司机结束行程并上报实际里程/时长/金额。需携带有效 JWT。
func FinishTripHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.FinishTripRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "orderId 不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).FinishTrip(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ConfirmArriveHandler POST /api/driver/v1/orders/confirm-arrive
// 当前登录司机确认已到达上车点。需携带有效 JWT。
func ConfirmArriveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.ConfirmArriveRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "orderId 不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).ConfirmArrive(int64(claims.AccountID), req.OrderID)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// RejectOrderHandler POST /api/driver/v1/orders/reject
// 当前登录司机拒绝派单。
func RejectOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.RejectOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "orderId 不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).RejectOrder(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListMyDispatchesHandler POST /api/driver/v1/orders/dispatches
// 查询当前登录司机的派单列表。
func ListMyDispatchesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.ListMyDispatchesRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Page < 0 || req.PageSize < 0 || req.Status < 0 {
			writeError(w, http.StatusBadRequest, 50000, "分页或状态参数不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).ListMyDispatches(
			int64(claims.AccountID),
			req.Page,
			req.PageSize,
			req.Status,
		)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListMyOrdersHandler POST /api/driver/v1/orders/list
// 直接从 ordersvc 查询当前登录司机的订单列表。
func ListMyOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.ListMyOrdersRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Page < 0 || req.PageSize < 0 || req.Status < 0 || req.Status > 6 {
			writeError(w, http.StatusBadRequest, 50000, "订单查询参数不合法")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).ListMyOrders(
			int64(claims.AccountID),
			req.Page,
			req.PageSize,
			req.Status,
		)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
