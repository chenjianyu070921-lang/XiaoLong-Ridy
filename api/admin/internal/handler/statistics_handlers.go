package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/admin/internal/logic"
)

// 统计处理器按指标域独立组织；路由、鉴权和响应格式与原实现保持一致。
func (r *Router) handleStatisticsOverview(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Overview(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
func (r *Router) handleStatisticsOrders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Orders(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
func (r *Router) handleStatisticsDrivers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Drivers(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
func (r *Router) handleStatisticsRevenue(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Revenue(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
func (r *Router) handleStatisticsCoupons(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Coupons(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
func (r *Router) handleStatisticsUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := logic.NewStatisticsLogic(r.ctx).Users(req.Context(), statisticsRequestFromQuery(req))
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, resp)
}
