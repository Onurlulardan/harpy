package iam

import "testing"

func TestResolveUsageAccessPinsScopePermissionAndOperations(t *testing.T) {
	for operation := range tenantUsageOperations {
		target, valid := resolveUsageAccess("tenant", "analytics.usage.read", operation)
		if !valid || target.resource != "analytics.usage" || target.action != "read" || target.platform {
			t.Fatalf("tenant Usage target for %s = %#v/%t", operation, target, valid)
		}
	}
	for operation := range platformUsageOperations {
		target, valid := resolveUsageAccess(
			"platform", "platform.analytics.usage.read", operation,
		)
		if !valid || !target.platform || target.capability != "platform.analytics.usage.read" {
			t.Fatalf("platform Usage target for %s = %#v/%t", operation, target, valid)
		}
	}
	for _, input := range [][3]string{
		{"tenant", "platform.analytics.usage.read", "usage.tenant.overview"},
		{"platform", "analytics.usage.read", "usage.platform.overview"},
		{"tenant", "analytics.usage.read", "usage.platform.overview"},
	} {
		if _, valid := resolveUsageAccess(input[0], input[1], input[2]); valid {
			t.Fatalf("Usage authority accepted %q", input)
		}
	}
}
