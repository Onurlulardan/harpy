package iam

import (
	"testing"
)

func TestPermissionRequestRoutesAreExact(t *testing.T) {
	routes := (&PermissionRequestHTTP{}).Routes()
	want := map[string]bool{
		"GET /api/permission-requests":                           true,
		"POST /api/permission-requests":                          true,
		"POST /api/permission-requests/bulk":                     true,
		"GET /api/permission-requests/mine":                      true,
		"GET /api/permission-requests/catalog/resources":         true,
		"GET /api/permission-requests/catalog/actions":           true,
		"GET /api/permission-requests/catalog/missing-resources": true,
		"GET /api/permission-requests/{id}":                      true,
		"POST /api/permission-requests/{id}/review":              true,
		"POST /api/permission-requests/{id}/approve":             true,
		"POST /api/permission-requests/{id}/reject":              true,
		"POST /api/permission-requests/{id}/cancel":              true,
		"POST /api/permission-requests/{id}/expire":              true,
	}
	if len(routes) != len(want) {
		t.Fatalf("route count = %d", len(routes))
	}
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		if !want[key] {
			t.Fatalf("unexpected route %s", key)
		}
	}
}
