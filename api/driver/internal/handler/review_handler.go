package handler

import (
	"net/http"
	"strconv"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// ListReceivedReviewsHandler 返回当前司机收到的乘客评价。
func ListReceivedReviewsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
		pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)
		req := &types.ListReviewsRequest{Page: int32(page), PageSize: int32(pageSize)}
		resp, err := logic.NewReviewLogic(r.Context(), svcCtx).ListReceivedReviews(int64(claims.AccountID), req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SubmitDriverReviewHandler 处理司机提交对乘客的评价请求。
func SubmitDriverReviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.SubmitDriverReviewRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewReviewLogic(r.Context(), svcCtx).SubmitDriverReview(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListGivenReviewsHandler 返回当前司机给出的乘客评价。
func ListGivenReviewsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
		pageSize, _ := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)
		req := &types.ListGivenReviewsRequest{Page: int32(page), PageSize: int32(pageSize)}
		resp, err := logic.NewReviewLogic(r.Context(), svcCtx).ListGivenReviews(int64(claims.AccountID), req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
