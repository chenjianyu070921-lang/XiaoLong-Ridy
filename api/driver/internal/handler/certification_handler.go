package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func CreateCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateCertificationRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).CreateCertification(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UpdateCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateCertificationRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).UpdateCertification(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func GetCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "认证ID不合法")
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).GetCertification(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func DeleteCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "认证ID不合法")
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).DeleteCertification(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func ListCertificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListCertificationsRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).ListCertifications(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
