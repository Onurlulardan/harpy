package api

import (
	"testing"

	billing "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestBuildBillingRouteGroupsPublishesCanonicalSurface(t *testing.T) {
	groups, err := buildBillingRouteGroups(new(billing.Handler))
	if err != nil {
		t.Fatal(err)
	}
	var routes []httpedge.Route
	for _, group := range groups {
		routes = append(routes, group...)
	}
	if len(groups) != 4 || len(routes) == 0 {
		t.Fatalf("Billing route groups=%d routes=%d", len(groups), len(routes))
	}
	if _, err := httpedge.NewRoutes(func(error) {}, routes...); err != nil {
		t.Fatalf("compile Billing routes: %v", err)
	}
}

func TestBuildBillingRouteGroupsFailsClosedOnMissingEdge(t *testing.T) {
	if _, err := buildBillingRouteGroups(nil); err == nil {
		t.Fatal("nil canonical Billing handler was accepted")
	}
}
