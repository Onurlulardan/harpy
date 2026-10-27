package iam

type usageAccessTarget struct {
	resource, action, capability string
	platform                     bool
}

var tenantUsageOperations = map[string]struct{}{
	"usage.tenant.overview":      {},
	"usage.tenant.active_users":  {},
	"usage.tenant.top_users":     {},
	"usage.tenant.user_summary":  {},
	"usage.tenant.user_timeline": {},
	"usage.tenant.features":      {},
	"usage.tenant.heatmap":       {},
	"usage.tenant.trend":         {},
}

var platformUsageOperations = map[string]struct{}{
	"usage.platform.overview":      {},
	"usage.platform.active_users":  {},
	"usage.platform.top_users":     {},
	"usage.platform.user_summary":  {},
	"usage.platform.user_timeline": {},
	"usage.platform.features":      {},
	"usage.platform.heatmap":       {},
	"usage.platform.trend":         {},
}

func resolveUsageAccess(scope, permission, operation string) (usageAccessTarget, bool) {
	if scope == "tenant" && permission == "analytics.usage.read" {
		_, valid := tenantUsageOperations[operation]
		return usageAccessTarget{resource: "analytics.usage", action: "read"}, valid
	}
	if scope == "platform" && permission == "platform.analytics.usage.read" {
		_, valid := platformUsageOperations[operation]
		return usageAccessTarget{
			capability: "platform.analytics.usage.read", platform: true,
		}, valid
	}
	return usageAccessTarget{}, false
}
