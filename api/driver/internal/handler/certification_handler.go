// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// CreateCertificationHandler POST /api/driver/v1/certifications
// 处理创建司机认证的 HTTP 请求。
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

// UpdateCertificationHandler POST /api/driver/v1/certifications/update
// 处理更新认证信息（含审核状态流转）的 HTTP 请求。
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

// GetCertificationHandler GET /api/driver/v1/certifications/get?id=
// 处理查询认证详情的 HTTP 请求（id 取自查询参数）。
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

// DeleteCertificationHandler POST /api/driver/v1/certifications/delete?id=
// 处理删除认证的 HTTP 请求（id 取自查询参数）。
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

// ListCertificationsHandler POST /api/driver/v1/certifications/list
// 处理分页查询认证列表的 HTTP 请求。
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
