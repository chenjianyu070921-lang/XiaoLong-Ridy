package handler

// registerDriverRoutes 注册司机、司机认证和提现路由。
func (r *Router) registerDriverRoutes() {
	r.mux.HandleFunc("/admin/v1/drivers", r.authRequired(r.handleDrivers))
	r.mux.HandleFunc("/admin/v1/drivers/", r.authRequired(r.handleDriverByID))
	r.mux.HandleFunc("/admin/v1/driver-withdrawals", r.authRequired(r.handleDriverWithdrawals))
	r.mux.HandleFunc("/admin/v1/driver-withdrawals/", r.authRequired(r.handleDriverWithdrawalByID))
	r.mux.HandleFunc("/admin/v1/driver-certifications", r.authRequired(r.handleDriverCertifications))
	r.mux.HandleFunc("/admin/v1/driver-certifications/", r.authRequired(r.handleDriverCertificationByID))
}
