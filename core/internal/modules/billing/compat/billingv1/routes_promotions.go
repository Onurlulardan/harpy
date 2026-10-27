package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

var adminPromotionRoutes = []Route{
	{"billing.route.admin-coupons-list", "GET", "/api/admin/billing/coupons", "/api/admin/billing/coupons", "platform.discounts.list", string(billingapi.PermissionPlatformBillingRead), MappingPreserved},
	{"billing.route.admin-coupons-create", "POST", "/api/admin/billing/coupons", "/api/admin/billing/coupons", "platform.discounts.create", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-coupon-update", "PUT", "/api/admin/billing/coupons/:id", "/api/admin/billing/coupons/{id}", "platform.discounts.version", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-coupon-delete", "DELETE", "/api/admin/billing/coupons/:id", "/api/admin/billing/coupons/{id}", "platform.discounts.retire", string(billingapi.PermissionPlatformBillingManage), MappingRetire},
	{"billing.route.admin-campaigns-list", "GET", "/api/admin/billing/campaigns", "/api/admin/billing/campaigns", "platform.campaigns.list", string(billingapi.PermissionPlatformBillingRead), MappingPreserved},
	{"billing.route.admin-campaigns-create", "POST", "/api/admin/billing/campaigns", "/api/admin/billing/campaigns", "platform.campaigns.create", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-campaign-update", "PUT", "/api/admin/billing/campaigns/:id", "/api/admin/billing/campaigns/{id}", "platform.campaigns.version", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-campaign-delete", "DELETE", "/api/admin/billing/campaigns/:id", "/api/admin/billing/campaigns/{id}", "platform.campaigns.retire", string(billingapi.PermissionPlatformBillingManage), MappingRetire},
}
