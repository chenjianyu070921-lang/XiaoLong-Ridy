package handler

// registerOperationRoutes 注册操作日志、导出、工单及风控路由。
func (r *Router) registerOperationRoutes() {
	r.mux.HandleFunc("/admin/v1/operation-logs", r.authRequired(r.handleOperationLogs))
	r.mux.HandleFunc("/admin/v1/export-tasks", r.authRequired(r.handleExportTasks))
	r.mux.HandleFunc("/admin/v1/export-tasks/", r.authRequired(r.handleExportTaskByNo))
	r.mux.HandleFunc("/admin/v1/work-orders", r.authRequired(r.handleWorkOrders))
	r.mux.HandleFunc("/admin/v1/work-orders/batch-actions", r.authRequired(r.handleWorkOrderBatchActions))
	r.mux.HandleFunc("/admin/v1/work-orders/", r.authRequired(r.handleWorkOrderByID))
	r.mux.HandleFunc("/admin/v1/blacklist", r.authRequired(r.handleBlacklists))
	r.mux.HandleFunc("/admin/v1/blacklist/", r.authRequired(r.handleBlacklistByID))
	r.mux.HandleFunc("/admin/v1/risk/hit-records", r.authRequired(r.handleRiskHitRecords))
	r.mux.HandleFunc("/admin/v1/risk/hit-records/actions", r.authRequired(r.handleRiskHitRecordActions))
	r.mux.HandleFunc("/admin/v1/notification-outbox", r.authRequired(r.handleNotificationOutbox))
}
