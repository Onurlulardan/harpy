package api

import (
	"fmt"

	billing "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	billingv1 "github.com/Onurlulardan/harpy/core/internal/modules/billing/compat/billingv1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func buildBillingRouteGroups(
	canonical *billing.Handler,
	compatibility *billingv1.Handler,
) ([][]httpedge.Route, error) {
	if canonical == nil || compatibility == nil {
		return nil, fmt.Errorf("canonical and compatibility Billing HTTP handlers are required")
	}
	bindings, err := compatibility.BindRoutes()
	if err != nil {
		return nil, fmt.Errorf("bind Billing V1 compatibility routes: %w", err)
	}
	legacy := make([]httpedge.Route, len(bindings))
	for index, binding := range bindings {
		legacy[index] = httpedge.Route{
			Method: binding.Method, Pattern: binding.Pattern, Handler: binding.Handler,
		}
	}
	return [][]httpedge.Route{
		canonical.PublicRoutes(),
		canonical.TenantRoutes(),
		canonical.PlatformRoutes(),
		canonical.WebhookRoutes(),
		legacy,
	}, nil
}
