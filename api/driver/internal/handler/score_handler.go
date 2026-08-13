// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// CreateScoreHandler POST /api/driver/v1/scores
// 处理创建司机服务分的 HTTP 请求。
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

// UpdateScoreHandler POST /api/driver/v1/scores/update
// 处理更新服务分的 HTTP 请求。
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

// GetScoreHandler GET /api/driver/v1/scores/get?id=
// 处理查询服务分详情的 HTTP 请求（id 取自查询参数）。
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

// DeleteScoreHandler POST /api/driver/v1/scores/delete?id=
// 处理删除服务分记录的 HTTP 请求（id 取自查询参数）。
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

// ListScoresHandler POST /api/driver/v1/scores/list
// 处理分页查询服务分列表的 HTTP 请求。
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
