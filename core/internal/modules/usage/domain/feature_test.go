package domain

import "testing"

func TestFeatureForRouteUsesTemplatesOnly(t *testing.T) {
	cases := map[string]string{
		"/api/auth/login":                         "Authentication",
		"/api/form-engine/dynamic/{form_slug}":    "Form: {form_slug}",
		"/api/form-engine/dynamic/purchase-order": "Form: purchase-order",
		"/api/traces/{id}":                        "Traces",
		"/api/system/health":                      "System Health",
		"/api/alerts":                             "System Health",
		"/api/usage-analytics/tenant/overview":    "Analytics",
		"/api/platform/operations/{id}":           "Platform Operations",
		"/unknown":                                "Other",
	}
	for route, expected := range cases {
		if actual := FeatureForRoute(route); actual != expected {
			t.Fatalf("%s = %s, want %s", route, actual, expected)
		}
	}
}
