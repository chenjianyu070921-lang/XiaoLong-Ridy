package handler

// registerUserRoutes 注册用户查询及用户券历史路由。
func (r *Router) registerUserRoutes() {
	r.mux.HandleFunc("/admin/v1/users", r.authRequired(r.handleUsers))
	r.mux.HandleFunc("/admin/v1/users/", r.authRequired(r.handleUserByID))
	r.mux.HandleFunc("/admin/v1/user-coupons", r.authRequired(r.handleUserCoupons))
}
