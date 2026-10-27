package licenseauthority

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

type Limit = billingapi.LicenseLimit

type EntitlementProjection = billingapi.LicenseEntitlementProjection

func ProjectEntitlements(snapshot VerifiedSnapshot) EntitlementProjection {
	features := snapshot.Features()
	limits := snapshot.Limits()
	return EntitlementProjection{
		Boolean: map[string]bool{
			"workflow.builder.enabled":            features.WorkflowBuilder,
			"api.builder.enabled":                 features.APIBuilder,
			"analytics.dashboard_builder.enabled": features.DashboardBuilder,
			"workflow.enabled":                    true,
			"api.access":                          true,
		},
		Limits: map[string]Limit{
			"iam.active_users":            projectLimit(limits.MaxUsers),
			"workflow.active_definitions": projectLimit(limits.MaxWorkflows),
			"api.active_endpoints":        projectLimit(limits.MaxAPIEndpoints),
			"analytics.active_dashboards": projectLimit(limits.MaxDashboards),
			"documents.storage_bytes":     {Unlimited: true},
		},
		Channel: snapshot.Channel(),
	}
}

func projectLimit(value int64) Limit {
	if value == 0 {
		return Limit{Unlimited: true}
	}
	return Limit{Value: value}
}
