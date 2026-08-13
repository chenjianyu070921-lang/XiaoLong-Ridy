// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// CreateWithdrawHandler POST /api/driver/v1/withdraws
// 处理创建提现申请的 HTTP 请求。
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

// UpdateWithdrawHandler POST /api/driver/v1/withdraws/update
// 处理更新提现记录（含打款状态流转）的 HTTP 请求。
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

// GetWithdrawHandler GET /api/driver/v1/withdraws/get?id=
// 处理查询提现详情的 HTTP 请求（id 取自查询参数）。
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

// DeleteWithdrawHandler POST /api/driver/v1/withdraws/delete?id=
// 处理删除提现记录的 HTTP 请求（id 取自查询参数）。
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

// ListWithdrawsHandler POST /api/driver/v1/withdraws/list
// 处理分页查询提现列表的 HTTP 请求。
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
