package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// ReportLocationHandler POST /api/driver/v1/drivers/location/report
// 司机位置上报：从 JWT 取司机 ID，解析 deviceID 与经纬度，调用逻辑刷新司机位置和在线保活状态。
// 返回在线状态、互踢标记与本次上报时间。需携带有效 JWT。
func ReportLocationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.ReportLocationRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewLocationLogic(r.Context(), svcCtx).ReportLocation(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
