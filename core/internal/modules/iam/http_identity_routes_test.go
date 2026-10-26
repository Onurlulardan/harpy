package iam

import (
	"net/http"
	"testing"
)

func TestIdentityHTTPRoutesAreExactTwentyFour(t *testing.T) {
	want := []string{
		http.MethodGet + " /api/users",
		http.MethodPost + " /api/users",
		http.MethodDelete + " /api/users",
		http.MethodPost + " /api/users/query",
		http.MethodGet + " /api/users/{id}",
		http.MethodPut + " /api/users/{id}",
		http.MethodDelete + " /api/users/{id}",
		http.MethodPatch + " /api/users/{id}/state",
		http.MethodGet + " /api/roles",
		http.MethodPost + " /api/roles",
		http.MethodDelete + " /api/roles",
		http.MethodPost + " /api/roles/query",
		http.MethodGet + " /api/roles/{id}",
		http.MethodPut + " /api/roles/{id}",
		http.MethodDelete + " /api/roles/{id}",
		http.MethodPatch + " /api/roles/{id}/state",
		http.MethodGet + " /api/organizations",
		http.MethodPost + " /api/organizations",
		http.MethodDelete + " /api/organizations",
		http.MethodPost + " /api/organizations/query",
		http.MethodGet + " /api/organizations/{id}",
		http.MethodPut + " /api/organizations/{id}",
		http.MethodDelete + " /api/organizations/{id}",
		http.MethodPatch + " /api/organizations/{id}/state",
	}
	routes := (&IdentityHTTP{}).Routes()
	if len(routes) != len(want) {
		t.Fatalf("route count=%d want=%d", len(routes), len(want))
	}
	seen := make(map[string]struct{}, len(routes))
	for index, route := range routes {
		actual := route.Method + " " + route.Pattern
		if actual != want[index] || route.Handler == nil {
			t.Fatalf("route[%d]=%q handler=%v want=%q", index, actual, route.Handler, want[index])
		}
		if _, duplicate := seen[actual]; duplicate {
			t.Fatalf("duplicate route %q", actual)
		}
		seen[actual] = struct{}{}
	}
	var handler *IdentityHTTP
	if routes := handler.Routes(); routes != nil {
		t.Fatalf("nil handler routes=%v", routes)
	}
}
