package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func CreateScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateScoreRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewScoreLogic(r.Context(), svcCtx).CreateScore(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UpdateScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateScoreRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewScoreLogic(r.Context(), svcCtx).UpdateScore(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func GetScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "服务分记录ID不合法")
			return
		}
		resp, err := logic.NewScoreLogic(r.Context(), svcCtx).GetScore(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func DeleteScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "服务分记录ID不合法")
			return
		}
		resp, err := logic.NewScoreLogic(r.Context(), svcCtx).DeleteScore(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func ListScoresHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListScoresRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewScoreLogic(r.Context(), svcCtx).ListScores(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
