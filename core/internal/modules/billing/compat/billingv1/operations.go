package billingv1

func (handler *Handler) operations() map[string]operation {
	return map[string]operation{
		"billing.route.public-plans":              handler.listPublicPlans,
		"billing.route.active-campaign":           handler.getActiveCampaign,
		"billing.route.subscription-read":         handler.getSubscription,
		"billing.route.usage-read":                handler.getUsage,
		"billing.route.change-plan":               handler.changePlan,
		"billing.route.cancel-subscription":       handler.cancelSubscription,
		"billing.route.resume-subscription":       handler.resumeSubscription,
		"billing.route.apply-coupon":              handler.applyCoupon,
		"billing.route.validate-coupon":           handler.validateCoupon,
		"billing.route.admin-plans-list":          handler.listPlans,
		"billing.route.admin-plans-create":        handler.createPlan,
		"billing.route.admin-plan-read":           handler.getPlan,
		"billing.route.admin-plan-update":         handler.updatePlan,
		"billing.route.admin-plan-delete":         handler.deletePlan,
		"billing.route.admin-subscriptions-list":  handler.listSubscriptions,
		"billing.route.admin-subscription-update": handler.updateSubscription,
		"billing.route.admin-events-list":         handler.listEvents,
		"billing.route.admin-revenue-read":        handler.getRevenue,
		"billing.route.admin-coupons-list":        handler.listCoupons,
		"billing.route.admin-coupons-create":      handler.createCoupon,
		"billing.route.admin-coupon-update":       handler.updateCoupon,
		"billing.route.admin-coupon-delete":       handler.deleteCoupon,
		"billing.route.admin-campaigns-list":      handler.listCampaigns,
		"billing.route.admin-campaigns-create":    handler.createCampaign,
		"billing.route.admin-campaign-update":     handler.updateCampaign,
		"billing.route.admin-campaign-delete":     handler.deleteCampaign,
	}
}
