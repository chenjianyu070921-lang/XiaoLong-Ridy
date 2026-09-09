package handler

// registerAuthRoutes 注册认证、管理员和菜单权限路由。
func (r *Router) registerAuthRoutes() {
	r.mux.HandleFunc("/admin/v1/auth/register", r.handleRegister)
	r.mux.HandleFunc("/admin/v1/auth/login", r.handleLogin)
	r.mux.HandleFunc("/admin/v1/auth/logout", r.authRequired(r.handleLogout))
	r.mux.HandleFunc("/admin/v1/auth/me", r.authRequired(r.handleMe))
	r.mux.HandleFunc("/admin/v1/menus", r.authRequired(r.handleMenus))
	r.mux.HandleFunc("/admin/v1/admins", r.authRequired(r.handleAdmins))
	r.mux.HandleFunc("/admin/v1/admins/", r.authRequired(r.handleAdminByID))
}
