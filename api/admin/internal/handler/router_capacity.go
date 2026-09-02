package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/admin/internal/logic"
)

// registerCapacityRoutes 注册后台实时运力地图路由。
func (r *Router) registerCapacityRoutes() {
	r.mux.HandleFunc("/admin/v1/capacity/map", r.authRequired(r.handleCapacityMap))
}

// handleCapacityMap 查询司机位置和运力状态快照。
func (r *Router) handleCapacityMap(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewCapacityLogic(r.ctx).Map(req.Context(), int32Query(req, "status", 0), int32Query(req, "online_status", 0), int32Query(req, "limit", 100))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
