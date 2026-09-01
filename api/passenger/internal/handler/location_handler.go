package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// ReverseGeocodeHandler 处理 POST /api/passenger/v1/location/reverse-geocode。
func ReverseGeocodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReverseGeocodeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewLocationLogic(r.Context(), svcCtx).ReverseGeocode(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GeocodeHandler 处理 POST /api/passenger/v1/location/geocode。
func GeocodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GeocodeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewLocationLogic(r.Context(), svcCtx).Geocode(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
