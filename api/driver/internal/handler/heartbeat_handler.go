package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// HeartbeatHandler POST /api/driver/v1/drivers/heartbeat
// 司机心跳上报：从 JWT 取司机 ID，解析 deviceID 与位置，调用逻辑刷新在线状态并判定多端互踢。
// 返回在线状态与是否被顶替（kicked=true 时前端需强制重新登录）。需携带有效 JWT。
func HeartbeatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.HeartbeatRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewHeartbeatLogic(r.Context(), svcCtx).Heartbeat(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
