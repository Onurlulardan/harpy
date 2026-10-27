package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

var routeOperations = map[string]string{
	"billing.route.subscription-read":         "billing.subscription.read",
	"billing.route.usage-read":                "billing.usage.read",
	"billing.route.change-plan":               "billing.subscription.change",
	"billing.route.cancel-subscription":       "billing.subscription.cancel",
	"billing.route.resume-subscription":       "billing.subscription.resume",
	"billing.route.apply-coupon":              "billing.discount.redeem",
	"billing.route.validate-coupon":           "billing.discount.validate",
	"billing.route.admin-plans-list":          "billing.catalog.offering.list",
	"billing.route.admin-plans-create":        "billing.catalog.offering.create",
	"billing.route.admin-plan-read":           "billing.catalog.offering.read",
	"billing.route.admin-plan-update":         "billing.catalog.offering.publish",
	"billing.route.admin-plan-delete":         "billing.catalog.offering.retire",
	"billing.route.admin-subscriptions-list":  "billing.report.subscription.list",
	"billing.route.admin-subscription-update": "billing.report.subscription.transition",
	"billing.route.admin-events-list":         "billing.report.audit.list",
	"billing.route.admin-revenue-read":        "billing.report.revenue.read",
	"billing.route.admin-coupons-list":        "billing.catalog.discount.list",
	"billing.route.admin-coupons-create":      "billing.catalog.discount.create",
	"billing.route.admin-coupon-update":       "billing.catalog.discount.publish",
	"billing.route.admin-coupon-delete":       "billing.catalog.discount.retire",
	"billing.route.admin-campaigns-list":      "billing.catalog.campaign.list",
	"billing.route.admin-campaigns-create":    "billing.catalog.campaign.create",
	"billing.route.admin-campaign-update":     "billing.catalog.campaign.publish",
	"billing.route.admin-campaign-delete":     "billing.catalog.campaign.retire",
}

func (handler *Handler) authorize(
	r *http.Request,
	routeID string,
	permission billingapi.Permission,
) (billingapi.Access, error) {
	operation, found := routeOperations[routeID]
	if !found {
		return billingapi.Access{}, billingapi.ErrForbidden
	}
	return handler.accessFor(r, permission, operation)
}

func (handler *Handler) catalogReadAccess(
	r *http.Request,
	operation string,
) (billingapi.Access, error) {
	return handler.accessFor(r, billingapi.PermissionPlatformBillingRead, operation)
}
