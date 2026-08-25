package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// GetProfileHandler 处理 POST /api/passenger/v1/profile/me。
func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetProfileRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewProfileLogic(r.Context(), svcCtx, bearerToken(r)).GetProfile(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// SubmitRealNameHandler 处理 POST /api/passenger/v1/profile/real-name。
func SubmitRealNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitRealNameRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewProfileLogic(r.Context(), svcCtx, bearerToken(r)).SubmitRealName(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// UpdateProfileHandler 处理 POST /api/passenger/v1/profile/update。
func UpdateProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateProfileRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewProfileLogic(r.Context(), svcCtx, bearerToken(r)).UpdateProfile(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
