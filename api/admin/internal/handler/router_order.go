package handler

// registerOrderRoutes 注册订单、异常订单及退款补偿路由。
func (r *Router) registerOrderRoutes() {
	r.mux.HandleFunc("/admin/v1/orders", r.authRequired(r.handleOrders))
	r.mux.HandleFunc("/admin/v1/orders/abnormal", r.authRequired(r.handleAbnormalOrders))
	r.mux.HandleFunc("/admin/v1/orders/", r.authRequired(r.handleOrderByID))
	r.mux.HandleFunc("/admin/v1/refund-retry-tasks", r.authRequired(r.handleRefundRetryTasks))
	r.mux.HandleFunc("/admin/v1/refund-retry-tasks/", r.authRequired(r.handleRefundRetryTaskByNo))
}
