package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// CreateOrderHandler 处理 POST /api/passenger/v1/orders/create。
func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).CreateOrder(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListOrdersHandler 处理 POST /api/passenger/v1/orders/list。
func ListOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListOrdersRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).ListOrders(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetOrderHandler 处理 POST /api/passenger/v1/orders/detail。
func GetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).GetOrder(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// CancelOrderHandler 处理 POST /api/passenger/v1/orders/cancel。
func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).CancelOrder(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// PayOrderHandler 处理 POST /api/passenger/v1/orders/pay。
func PayOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).PayOrder(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetPaymentStatusHandler 处理 POST /api/passenger/v1/orders/payment-status。
func GetPaymentStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PaymentStatusRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).GetPaymentStatus(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetDispatchStatusHandler 处理 POST /api/passenger/v1/orders/dispatch-status。
func GetDispatchStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DispatchStatusRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).GetDispatchStatus(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// PollOrderStatusHandler 处理乘客端订单状态轮询兜底请求。
func PollOrderStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderStatusPollRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).PollOrderStatus(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// EstimateOrderHandler 处理 POST /api/passenger/v1/orders/estimate，用于下单前查询预估价格。
func EstimateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EstimateOrderRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOrderLogic(r.Context(), svcCtx, bearerToken(r)).EstimateOrder(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
