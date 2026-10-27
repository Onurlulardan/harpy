package iam

import "testing"

func TestResolveAnalyticsAccessAllowsDashboardDefaultReadOperation(t *testing.T) {
	target, ok := resolveAnalyticsAccess(
		"analytics.dashboard.read",
		"analytics.dashboard.read",
	)
	if !ok {
		t.Fatal("dashboard default read operation must be accepted by the read permission")
	}
	if target.resource != "analytics.dashboard" || target.action != "read" {
		t.Fatalf("unexpected target: %#v", target)
	}
}

func TestResolveAnalyticsAccessRejectsUnrelatedDashboardReadOperation(t *testing.T) {
	if _, ok := resolveAnalyticsAccess(
		"analytics.dashboard.read",
		"analytics.dashboard.update",
	); ok {
		t.Fatal("dashboard read permission must not accept update operation")
	}
}

func TestResolveAnalyticsAccessAllowsPublicationCreateAudiencePermissions(t *testing.T) {
	for _, permission := range []string{
		"analytics.publication.publish_public",
		"analytics.publication.publish_capability",
	} {
		t.Run(permission, func(t *testing.T) {
			target, ok := resolveAnalyticsAccess(
				permission,
				"analytics.publication.create",
			)
			if !ok {
				t.Fatalf("%s must accept publication create", permission)
			}
			if target.resource != "analytics.publication" {
				t.Fatalf("unexpected target: %#v", target)
			}
		})
	}
}

func TestResolveAnalyticsAccessPinsDashboardExportOperations(t *testing.T) {
	operations := []string{
		"analytics.dashboard.export.create",
		"analytics.dashboard.export.get",
		"analytics.dashboard.export.delete",
		"analytics.dashboard.export.content",
	}
	for _, operation := range operations {
		target, ok := resolveAnalyticsAccess("analytics.dashboard.export", operation)
		if !ok || target.resource != "analytics.dashboard" || target.action != "export" {
			t.Fatalf("dashboard export target for %s = %#v/%t", operation, target, ok)
		}
	}
	for _, operation := range []string{
		"analytics.dashboard.read", "analytics.dashboard.export", "analytics.query.execute",
	} {
		if _, ok := resolveAnalyticsAccess("analytics.dashboard.export", operation); ok {
			t.Fatalf("dashboard export permission accepted %s", operation)
		}
	}
}
