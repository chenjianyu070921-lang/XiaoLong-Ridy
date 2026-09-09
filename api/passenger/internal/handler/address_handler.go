package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
)

// CreateAddressHandler 处理 POST /api/passenger/v1/addresses/create。
func CreateAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateAddressRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAddressLogic(r.Context(), svcCtx, bearerToken(r)).CreateAddress(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListAddressesHandler 处理 POST /api/passenger/v1/addresses/list。
func ListAddressesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAddressesRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAddressLogic(r.Context(), svcCtx, bearerToken(r)).ListAddresses(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// UpdateAddressHandler 处理 POST /api/passenger/v1/addresses/update。
func UpdateAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateAddressRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAddressLogic(r.Context(), svcCtx, bearerToken(r)).UpdateAddress(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// DeleteAddressHandler 处理 POST /api/passenger/v1/addresses/delete。
func DeleteAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteAddressRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAddressLogic(r.Context(), svcCtx, bearerToken(r)).DeleteAddress(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
