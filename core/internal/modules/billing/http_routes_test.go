package billing

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestCanonicalRouteGroupsAreExact(t *testing.T) {
	handler := newHandlerFixture(t, nil, nil)
	groups := []struct {
		name  string
		count int
		got   int
	}{
		{name: "public", count: 2, got: len(handler.PublicRoutes())},
		{name: "tenant", count: 7, got: len(handler.TenantRoutes())},
		{name: "platform", count: 24, got: len(handler.PlatformRoutes())},
		{name: "webhook", count: 1, got: len(handler.WebhookRoutes())},
	}
	seen := make(map[string]struct{})
	for _, group := range groups {
		if group.got != group.count {
			t.Fatalf("%s route count = %d, want %d", group.name, group.got, group.count)
		}
	}
	for _, routes := range [][]httpedge.Route{
		handler.PublicRoutes(), handler.TenantRoutes(),
		handler.PlatformRoutes(), handler.WebhookRoutes(),
	} {
		for _, route := range routes {
			key := route.Method + " " + route.Pattern
			if _, duplicate := seen[key]; duplicate {
				t.Fatalf("duplicate route %s", key)
			}
			seen[key] = struct{}{}
		}
	}
	if len(seen) != 34 {
		t.Fatalf("canonical route total = %d", len(seen))
	}
}
