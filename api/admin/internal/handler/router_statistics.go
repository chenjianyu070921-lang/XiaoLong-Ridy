package handler

// registerStatisticsRoutes 注册统计查询路由。
func (r *Router) registerStatisticsRoutes() {
	r.mux.HandleFunc("/admin/v1/statistics/overview", r.authRequired(r.handleStatisticsOverview))
	r.mux.HandleFunc("/admin/v1/statistics/orders", r.authRequired(r.handleStatisticsOrders))
	r.mux.HandleFunc("/admin/v1/statistics/drivers", r.authRequired(r.handleStatisticsDrivers))
	r.mux.HandleFunc("/admin/v1/statistics/revenue", r.authRequired(r.handleStatisticsRevenue))
	r.mux.HandleFunc("/admin/v1/statistics/coupons", r.authRequired(r.handleStatisticsCoupons))
	r.mux.HandleFunc("/admin/v1/statistics/users", r.authRequired(r.handleStatisticsUsers))
}
