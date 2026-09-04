package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func CreateVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.CreateVehicleRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).CreateVehicle(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UpdateVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.UpdateVehicleRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).UpdateVehicle(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func GetVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "车辆ID不合法")
			return
		}
		driverID := int64(0)
		if middleware.IsInternalCall(r.Context()) {
			var ok bool
			driverID, ok = decodeQueryID(r, "driverId")
			if !ok {
				writeError(w, http.StatusBadRequest, 50000, "司机ID不合法")
				return
			}
		} else {
			claims := middleware.ClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
				return
			}
			driverID = int64(claims.AccountID)
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).GetVehicle(driverID, id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func DeleteVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		id, ok := decodeQueryID(r, "id")
		if !ok {
			var req types.DeleteVehicleRequest
			if !decodeJSON(w, r, &req) {
				return
			}
			id = req.ID
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).DeleteVehicle(int64(claims.AccountID), id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
