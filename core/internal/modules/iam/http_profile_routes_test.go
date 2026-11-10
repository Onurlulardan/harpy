package iam

import (
	"net/http"
	"testing"
)

func TestProfileHTTPRoutesAreSelfOnly(t *testing.T) {
	want := []string{
		http.MethodGet + " /api/profile",
		http.MethodPut + " /api/profile",
	}
	routes := (&ProfileHTTP{}).Routes()
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
	var handler *ProfileHTTP
	if routes := handler.Routes(); routes != nil {
		t.Fatalf("nil handler routes=%v", routes)
	}
}
