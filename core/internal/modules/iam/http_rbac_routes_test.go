package iam

import "testing"

func TestRBACRoutesAreExact(t *testing.T) {
	routes := (&RBACHTTP{}).Routes()
	want := map[string]bool{
		"GET /api/permissions": true, "POST /api/permissions": true,
		"GET /api/permissions/{id}": true, "PUT /api/permissions/{id}": true,
		"DELETE /api/permissions/{id}":   true,
		"GET /api/permissions/resources": true, "POST /api/permissions/resources": true,
		"GET /api/permissions/resources/{id}":    true,
		"PUT /api/permissions/resources/{id}":    true,
		"DELETE /api/permissions/resources/{id}": true,
		"GET /api/permissions/actions":           true, "POST /api/permissions/actions": true,
		"GET /api/permissions/actions/{id}":                         true,
		"PUT /api/permissions/actions/{id}":                         true,
		"DELETE /api/permissions/actions/{id}":                      true,
		"POST /api/permissions/check":                               true,
		"POST /api/permissions/batch-check":                         true,
		"GET /api/users/{id}/permissions":                           true,
		"GET /api/roles/{id}/permissions":                           true,
		"GET /api/organizations/{id}/permissions":                   true,
		"GET /api/permissions/cache/stats":                          true,
		"POST /api/permissions/cache/invalidate/{user_id}":          true,
		"POST /api/permissions/cache/invalidate/role/{role_id}":     true,
		"POST /api/permissions/cache/invalidate/org/{org_id}":       true,
		"POST /api/permissions/cache/invalidate/tenant/{tenant_id}": true,
		"POST /api/permissions/cache/invalidate/all":                true,
	}
	if len(routes) != len(want) {
		t.Fatalf("route count = %d, want %d", len(routes), len(want))
	}
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if !want[key] {
			t.Fatalf("unexpected route %s", key)
		}
	}
}
