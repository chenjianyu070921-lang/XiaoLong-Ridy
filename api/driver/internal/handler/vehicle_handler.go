package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func CreateVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateVehicleRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).CreateVehicle(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UpdateVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateVehicleRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).UpdateVehicle(&req)
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
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).GetVehicle(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func DeleteVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "车辆ID不合法")
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).DeleteVehicle(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func ListVehiclesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListVehiclesRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).ListVehicles(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
