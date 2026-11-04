package api

import (
	"fmt"

	billing "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func buildBillingRouteGroups(canonical *billing.Handler) ([][]httpedge.Route, error) {
	if canonical == nil {
		return nil, fmt.Errorf("canonical Billing HTTP handler is required")
	}
	return [][]httpedge.Route{
		canonical.PublicRoutes(),
		canonical.TenantRoutes(),
		canonical.PlatformRoutes(),
		canonical.WebhookRoutes(),
	}, nil
}
