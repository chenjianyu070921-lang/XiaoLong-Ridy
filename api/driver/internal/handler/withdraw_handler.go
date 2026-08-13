package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func CreateWithdrawHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateWithdrawRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewWithdrawLogic(r.Context(), svcCtx).CreateWithdraw(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UpdateWithdrawHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateWithdrawRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewWithdrawLogic(r.Context(), svcCtx).UpdateWithdraw(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func GetWithdrawHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "提现记录ID不合法")
			return
		}
		resp, err := logic.NewWithdrawLogic(r.Context(), svcCtx).GetWithdraw(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func DeleteWithdrawHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "提现记录ID不合法")
			return
		}
		resp, err := logic.NewWithdrawLogic(r.Context(), svcCtx).DeleteWithdraw(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func ListWithdrawsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListWithdrawsRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewWithdrawLogic(r.Context(), svcCtx).ListWithdraws(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
