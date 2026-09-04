package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// AcceptOrderHandler POST /api/driver/v1/orders/accept
// Current driver accepts an order. Requires a valid JWT.
func AcceptOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.AcceptOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
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

// CancelOrderHandler POST /api/driver/v1/orders/cancel
// Current driver cancels an accepted order before trip start.
func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.CancelOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).CancelOrder(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// StartTripHandler POST /api/driver/v1/orders/start-trip
// Current driver starts the trip.
func StartTripHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.StartTripRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
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
// Current driver finishes the trip and reports final trip metrics.
func FinishTripHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.FinishTripRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
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
// Current driver confirms arrival at pickup point.
func ConfirmArriveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.ConfirmArriveRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
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
// Current driver rejects a dispatched order.
func RejectOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.RejectOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.OrderID <= 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid orderId")
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
// Lists dispatch records for the current driver.
func ListMyDispatchesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMyDispatchesRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Page < 0 || req.PageSize < 0 || req.Status < 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid dispatch query parameters")
			return
		}
		driverID := int64(0)
		if middleware.IsInternalCall(r.Context()) {
			if req.DriverID <= 0 {
				writeError(w, http.StatusBadRequest, 50000, "invalid driverId")
				return
			}
			driverID = req.DriverID
		} else {
			claims := middleware.ClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
				return
			}
			driverID = int64(claims.AccountID)
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).ListMyDispatches(
			driverID,
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

// ListAvailableOrdersHandler POST /api/driver/v1/orders/available
// Lists nearby WAIT_ACCEPT orders for the current driver.
func ListAvailableOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.ListMyOrdersRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Page < 0 || req.PageSize < 0 {
			writeError(w, http.StatusBadRequest, 50000, "invalid order query parameters")
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).ListAvailableOrders(int64(claims.AccountID), req.Page, req.PageSize)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListMyOrdersHandler POST /api/driver/v1/orders/list
// Lists orders assigned to the current driver.
func ListMyOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.ListMyOrdersRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Page < 0 || req.PageSize < 0 || req.Status < 0 || req.Status > 7 {
			writeError(w, http.StatusBadRequest, 50000, "invalid order query parameters")
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

func GetMyOrderDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.GetMyOrderDetailRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx).GetMyOrderDetail(int64(claims.AccountID), req.OrderID)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
