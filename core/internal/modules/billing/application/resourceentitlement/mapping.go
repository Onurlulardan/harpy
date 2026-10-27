package resourceentitlement

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

type resourceSpec struct {
	feature entitlement.Key
	gauge   entitlement.Key
}

func resolveSpec(kind billingapi.MeteredResourceKind) (resourceSpec, error) {
	var featureRaw, gaugeRaw string
	switch kind {
	case billingapi.MeteredResourceAPIEndpoint:
		featureRaw, gaugeRaw = "api.builder.enabled", "api.active_endpoints"
	case billingapi.MeteredResourceAnalyticsDashboard:
		featureRaw = "analytics.dashboard_builder.enabled"
		gaugeRaw = "analytics.active_dashboards"
	default:
		return resourceSpec{}, ErrRequest
	}
	feature, firstErr := entitlement.NewKey(featureRaw)
	gauge, secondErr := entitlement.NewKey(gaugeRaw)
	if firstErr != nil || secondErr != nil {
		return resourceSpec{}, ErrConfiguration
	}
	return resourceSpec{feature: feature, gauge: gauge}, nil
}
