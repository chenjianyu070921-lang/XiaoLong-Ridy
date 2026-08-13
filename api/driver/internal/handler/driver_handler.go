package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// CreateDriverHandler POST /api/driver/v1/drivers
func CreateDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateDriverRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).CreateDriver(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// UpdateDriverHandler POST /api/driver/v1/drivers/update
func UpdateDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateDriverRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).UpdateDriver(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetDriverHandler GET /api/driver/v1/drivers/:id
func GetDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "司机ID不合法")
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).GetDriver(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// DeleteDriverHandler POST /api/driver/v1/drivers/delete
func DeleteDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "司机ID不合法")
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).DeleteDriver(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListDriversHandler POST /api/driver/v1/drivers/list
func ListDriversHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListDriversRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).ListDrivers(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
