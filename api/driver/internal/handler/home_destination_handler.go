package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// GetHomeDestinationHandler GET /api/driver/v1/home-destination
// 查询当前司机的回家目的地与回家顺路模式状态（未设置时 hasSetting=false）。
func GetHomeDestinationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		resp, err := logic.NewHomeDestinationLogic(r.Context(), svcCtx).Get(int64(claims.AccountID))
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SetHomeDestinationHandler POST /api/driver/v1/home-destination
// 保存回家目的地；open=true 时同时开启回家顺路模式（仅做订单顺路过滤，不改变听单状态）。
func SetHomeDestinationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.SetHomeDestinationRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewHomeDestinationLogic(r.Context(), svcCtx).Set(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SetHomeModeHandler POST /api/driver/v1/home-destination/mode
// 仅切换回家顺路模式开关：开启=顺路过滤，关闭=恢复全域听单。
func SetHomeModeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.SetHomeModeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewHomeDestinationLogic(r.Context(), svcCtx).SetMode(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
