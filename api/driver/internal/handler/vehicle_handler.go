// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// CreateVehicleHandler POST /api/driver/v1/vehicles
// 处理创建车辆的 HTTP 请求。
func CreateVehicleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 声明车辆创建请求并解析 JSON 体。
		var req types.CreateVehicleRequest
		if !decodeJSON(w, r, &req) {
			return // 解析失败已写响应
		}
		// 调用业务逻辑层创建车辆。
		resp, err := logic.NewVehicleLogic(r.Context(), svcCtx).CreateVehicle(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// UpdateVehicleHandler POST /api/driver/v1/vehicles/update
// 处理更新车辆信息的 HTTP 请求。
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

// GetVehicleHandler GET /api/driver/v1/vehicles/get?id=
// 处理查询车辆详情的 HTTP 请求（id 取自查询参数）。
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

// DeleteVehicleHandler POST /api/driver/v1/vehicles/delete?id=
// 处理删除车辆的 HTTP 请求（id 取自查询参数）。
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

// ListVehiclesHandler POST /api/driver/v1/vehicles/list
// 处理分页查询车辆列表的 HTTP 请求。
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
