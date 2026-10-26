package iam

import (
	"net/http"
	"testing"
)

func TestPortalUserHTTPRoutesAreExactFive(t *testing.T) {
	want := []string{
		http.MethodPost + " /api/auth/portal-users",
		http.MethodPost + " /api/auth/portal-users/{id}/invite",
		http.MethodPatch + " /api/auth/portal-users/{id}/status",
		http.MethodPost + " /api/auth/portal-users/{id}/console-access",
		http.MethodDelete + " /api/auth/portal-users/{id}/console-access",
	}
	routes := (&PortalUserHTTP{}).Routes()
	if len(routes) != len(want) {
		t.Fatalf("route count=%d want=%d", len(routes), len(want))
	}
	seen := make(map[string]bool, len(routes))
	for index, route := range routes {
		actual := route.Method + " " + route.Pattern
		if actual != want[index] || route.Handler == nil || seen[actual] {
			t.Fatalf("route[%d]=%q handler=%v duplicate=%t", index,
				actual, route.Handler, seen[actual])
		}
		seen[actual] = true
	}
	var handler *PortalUserHTTP
	if handler.Routes() != nil {
		t.Fatal("nil portal user handler exposed routes")
	}
}
