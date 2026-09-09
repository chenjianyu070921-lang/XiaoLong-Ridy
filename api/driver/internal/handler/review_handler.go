package handler

import (
	"net/http"
	"strconv"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// ReviewSummaryHandler 返回当前司机的评价概览（含 agent 生成的服务平均分）。
func ReviewSummaryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		resp, err := logic.NewReviewLogic(r.Context(), svcCtx).GetReviewSummary(int64(claims.AccountID))
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

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
