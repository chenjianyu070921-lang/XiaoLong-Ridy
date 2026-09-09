package handler

// registerPunishmentRoutes 注册处罚规则、处罚单和处罚申诉路由。
func (r *Router) registerPunishmentRoutes() {
	r.mux.HandleFunc("/admin/v1/punishment-rules", r.authRequired(r.handlePunishmentRules))
	r.mux.HandleFunc("/admin/v1/punishment-rules/", r.authRequired(r.handlePunishmentRuleByID))
	r.mux.HandleFunc("/admin/v1/punishments", r.authRequired(r.handlePunishments))
	r.mux.HandleFunc("/admin/v1/punishments/", r.authRequired(r.handlePunishmentByID))
	r.mux.HandleFunc("/admin/v1/punishment-appeals", r.authRequired(r.handlePunishmentAppeals))
	r.mux.HandleFunc("/admin/v1/punishment-appeals/", r.authRequired(r.handlePunishmentAppealByID))
}
