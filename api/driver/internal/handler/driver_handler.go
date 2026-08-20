// Package handler 实现司机端 HTTP API 的路由与请求处理。
package handler

import (
	"net/http" // HTTP 处理器基础类型

	"XiaoLong-Ridy/api/driver/internal/logic" // 业务逻辑层
	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文
	"XiaoLong-Ridy/api/driver/internal/types"  // 请求/响应类型
)

// RegisterDriverHandler POST /api/driver/v1/drivers/register
// 司机自注册入口：解析请求、调用司机逻辑注册账号（公开接口，无需登录，初始状态待审核），成功返回司机 ID 与状态。
func RegisterDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 返回闭包处理器。
	return func(w http.ResponseWriter, r *http.Request) {
		// 声明请求结构体并解析 JSON 体。
		var req types.RegisterDriverRequest
		if !decodeJSON(w, r, &req) {
			return // 解析失败已写响应，直接返回
		}
		// 调用业务逻辑层注册司机。
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).RegisterDriver(&req)
		if err != nil {
			// 校验/下游错误，映射为统一错误响应。
			writeParamError(w, err)
			return
		}
		// 成功，写回响应。
		writeSuccess(w, resp)
	}
}

// CreateDriverHandler POST /api/driver/v1/drivers
// 管理员创建司机（后台建号）：需登录鉴权，解析请求、调用司机逻辑创建账号，成功返回司机 ID 与状态。
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
// 更新司机信息：解析请求、调用司机逻辑转发可选字段到 driversvc，成功返回更新后的司机 ID 与状态。
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
// 查询司机详情：从查询参数解析 id（非法返回 400），调用司机逻辑查询并返回脱敏后的司机资料。
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
// 删除（软删）司机：从查询参数解析 id（非法返回 400），调用司机逻辑软删并返回是否成功。
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

// GetDriverAiScoreHandler GET /api/driver/v1/drivers/ai-score?id=
// 查询司机 AI 智能推荐得分：从查询参数解析司机 id（非法返回 400），调用司机逻辑获取综合推荐分与影响因子，返回降级标记。需携带有效 JWT。
func GetDriverAiScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从查询参数解析司机 id。
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "司机ID不合法")
			return
		}
		// 调用司机逻辑查询 AI 推荐得分（含降级透传）。
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).GetDriverAiScore(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
