// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// CreateDriverHandler POST /api/driver/v1/drivers
// 处理创建司机的 HTTP 请求。
func CreateDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 返回闭包处理器。
	return func(w http.ResponseWriter, r *http.Request) {
		// 声明请求结构体并解析 JSON 体。
		var req types.CreateDriverRequest
		if !decodeJSON(w, r, &req) {
			return // 解析失败已写响应，直接返回
		}
		// 调用业务逻辑层创建司机。
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).CreateDriver(&req)
		if err != nil {
			// 校验/下游错误，映射为统一错误响应。
			writeParamError(w, err)
			return
		}
		// 成功，写回响应。
		writeSuccess(w, resp)
	}
}

// UpdateDriverHandler POST /api/driver/v1/drivers/update
// 处理更新司机信息的 HTTP 请求。
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

// GetDriverHandler GET /api/driver/v1/drivers/get?id=
// 处理查询司机详情的 HTTP 请求（id 取自查询参数）。
func GetDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从查询参数解析 id。
		id, ok := decodeQueryID(r, "id")
		if !ok {
			// id 非法，返回参数错误。
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

// DeleteDriverHandler POST /api/driver/v1/drivers/delete?id=
// 处理删除（软删）司机的 HTTP 请求（id 取自查询参数）。
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
