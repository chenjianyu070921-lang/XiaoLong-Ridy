package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// SetOnlineHandler POST /api/driver/v1/drivers/online
// 司机上线：从 JWT 取司机 ID，解析上报位置，调用逻辑将司机置为在线并写入位置，返回在线状态。需携带有效 JWT。
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
		resp, err := logic.NewOnlineLogic(r.Context(), svcCtx).SetOnline(int64(claims.AccountID), req.DeviceID, req.Longitude, req.Latitude)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SetOfflineHandler POST /api/driver/v1/drivers/offline
// 司机下线：从 JWT 取司机 ID，解析上报位置，调用逻辑将司机置为离线并更新位置，返回在线状态。需携带有效 JWT。
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
		resp, err := logic.NewOfflineLogic(r.Context(), svcCtx).SetOffline(int64(claims.AccountID), req.DeviceID, req.Longitude, req.Latitude)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
