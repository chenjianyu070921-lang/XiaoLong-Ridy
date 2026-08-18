package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// ClaimCouponHandler 处理 POST /api/passenger/v1/coupons/claim。
func ClaimCouponHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClaimCouponRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewCouponLogic(r.Context(), svcCtx, bearerToken(r)).ClaimCoupon(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListMyCouponsHandler 处理 POST /api/passenger/v1/coupons/my。
func ListMyCouponsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMyCouponsRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewCouponLogic(r.Context(), svcCtx, bearerToken(r)).ListMyCoupons(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
