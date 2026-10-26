package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

var publicTenantRoutes = []Route{
	{"billing.route.public-plans", "GET", "/api/billing/plans", "/api/billing/plans", "public.offerings.list", "none", MappingPreserved},
	{"billing.route.active-campaign", "GET", "/api/billing/campaigns/active", "/api/billing/campaigns/active", "public.campaign.active", "none", MappingPreserved},
	{"billing.route.subscription-read", "GET", "/api/billing/subscription", "/api/billing/subscription", "tenant.subscription.current", string(billingapi.PermissionBillingRead), MappingPreserved},
	{"billing.route.usage-read", "GET", "/api/billing/usage", "/api/billing/usage", "tenant.usage.current", string(billingapi.PermissionBillingUsageRead), MappingPreserved},
	{"billing.route.change-plan", "POST", "/api/billing/subscription/change-plan", "/api/billing/subscription/change-plan", "tenant.subscription.change", string(billingapi.PermissionBillingManage), MappingVersion},
	{"billing.route.cancel-subscription", "POST", "/api/billing/subscription/cancel", "/api/billing/subscription/cancel", "tenant.subscription.cancel", string(billingapi.PermissionBillingManage), MappingTransition},
	{"billing.route.resume-subscription", "POST", "/api/billing/subscription/resume", "/api/billing/subscription/resume", "tenant.subscription.resume", string(billingapi.PermissionBillingManage), MappingTransition},
	{"billing.route.apply-coupon", "POST", "/api/billing/subscription/apply-coupon", "/api/billing/subscription/apply-coupon", "tenant.discount.redeem", string(billingapi.PermissionBillingManage), MappingPreserved},
	{"billing.route.validate-coupon", "POST", "/api/billing/coupons/validate", "/api/billing/coupons/validate", "tenant.discount.validate", string(billingapi.PermissionBillingRead), MappingPreserved},
}
