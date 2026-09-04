package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// SetOnlineHandler POST /api/driver/v1/drivers/online
// 当前登录司机上线（置为在线状态）。需携带有效 JWT。
func SetOnlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.SetOnlineRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOnlineLogic(r.Context(), svcCtx).SetOnline(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SetOfflineHandler POST /api/driver/v1/drivers/offline
// 当前登录司机下线（置为离线状态）。需携带有效 JWT。
func SetOfflineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.SetOfflineRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewOfflineLogic(r.Context(), svcCtx).SetOffline(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
