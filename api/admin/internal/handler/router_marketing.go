package handler

// registerMarketingRoutes 注册优惠券、计价规则和营销活动路由。
func (r *Router) registerMarketingRoutes() {
	r.mux.HandleFunc("/admin/v1/coupons", r.authRequired(r.handleCoupons))
	r.mux.HandleFunc("/admin/v1/coupon-issue-tasks", r.authRequired(r.handleCouponIssueTasks))
	r.mux.HandleFunc("/admin/v1/coupons/", r.authRequired(r.handleCouponByID))
	r.mux.HandleFunc("/admin/v1/price-rules", r.authRequired(r.handlePriceRules))
	r.mux.HandleFunc("/admin/v1/price-rules/", r.authRequired(r.handlePriceRuleByID))
	r.mux.HandleFunc("/admin/v1/promotion-activities", r.authRequired(r.handlePromotionActivities))
	r.mux.HandleFunc("/admin/v1/promotion-activities/", r.authRequired(r.handlePromotionActivityByID))
}
