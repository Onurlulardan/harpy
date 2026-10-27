package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

var adminPlanRoutes = []Route{
	{"billing.route.admin-plans-list", "GET", "/api/admin/billing/plans", "/api/admin/billing/plans", "platform.offerings.list", string(billingapi.PermissionPlatformBillingRead), MappingPreserved},
	{"billing.route.admin-plans-create", "POST", "/api/admin/billing/plans", "/api/admin/billing/plans", "platform.offerings.create", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-plan-read", "GET", "/api/admin/billing/plans/:id", "/api/admin/billing/plans/{id}", "platform.offerings.get", string(billingapi.PermissionPlatformBillingRead), MappingPreserved},
	{"billing.route.admin-plan-update", "PUT", "/api/admin/billing/plans/:id", "/api/admin/billing/plans/{id}", "platform.offerings.version", string(billingapi.PermissionPlatformBillingManage), MappingVersion},
	{"billing.route.admin-plan-delete", "DELETE", "/api/admin/billing/plans/:id", "/api/admin/billing/plans/{id}", "platform.offerings.retire", string(billingapi.PermissionPlatformBillingManage), MappingRetire},
}

var adminReportRoutes = []Route{
	{"billing.route.admin-subscriptions-list", "GET", "/api/admin/billing/subscriptions", "/api/admin/billing/subscriptions", "platform.subscriptions.list", string(billingapi.PermissionPlatformBillingRead), MappingPreserved},
	{"billing.route.admin-subscription-update", "PUT", "/api/admin/billing/subscriptions/:id", "/api/admin/billing/subscriptions/{id}", "platform.subscriptions.transition", string(billingapi.PermissionPlatformBillingManage), MappingTransition},
	{"billing.route.admin-events-list", "GET", "/api/admin/billing/events", "/api/admin/billing/events", "platform.events.list", string(billingapi.PermissionPlatformBillingAudit), MappingPreserved},
	{"billing.route.admin-revenue-read", "GET", "/api/admin/billing/revenue", "/api/admin/billing/revenue", "platform.revenue.query", string(billingapi.PermissionPlatformBillingRevenue), MappingPreserved},
}
