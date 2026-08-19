package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// AcceptOrderHandler POST /api/driver/v1/orders/accept
// 职责：处理「司机接单」请求。
// 流程：① 从 JWT 取出当前登录司机 driverID（无凭证返回 401）；
//       ② 解析请求体取 orderId，校验 >0（非法返回 400）；
//       ③ 调用 OrderLogic.AcceptOrder，经 ordersvc 将订单由「待接单(1)」推进到「已接单(2)」；
//       ④ 将 ordersvc 返回的结果写入统一响应。
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
// 职责：处理「司机开始行程」请求。
// 流程：① 从 JWT 取出当前登录司机 driverID（无凭证返回 401）；
//       ② 解析请求体取 orderId，校验 >0（非法返回 400）；
//       ③ 调用 OrderLogic.StartTrip，经 ordersvc 将订单由「已接单(2)」推进到「行程中(3)」；
//       ④ 将 ordersvc 返回的结果写入统一响应。
// 注意：仅「已接单」状态的订单可开始行程，越级调用会被 ordersvc 拒绝。
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
// 职责：处理「司机结束行程」请求，并上报实际里程/时长/金额。
// 流程：① 从 JWT 取出当前登录司机 driverID（无凭证返回 401）；
//       ② 解析请求体取 orderId 与实际上报数据（里程/时长/金额），校验 orderId >0（非法返回 400）；
//       ③ 调用 OrderLogic.FinishTrip，经 ordersvc 将订单由「行程中(3)」推进到「待支付(4)」；
//       ④ 将 ordersvc 返回的订单状态与应付金额写入统一响应。
// 注意：仅「行程中」状态的订单可结束行程，越级调用会被 ordersvc 拒绝。
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
// 确认到达上车点：从 JWT 取司机 ID，解析 orderId，调用逻辑通知 ordersvc 记录"司机已到达"（订单状态不变，仍为已接单）。需携带有效 JWT。
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
