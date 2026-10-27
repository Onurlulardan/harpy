package api

import (
	"testing"

	billing "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	billingv1 "github.com/Onurlulardan/harpy/core/internal/modules/billing/compat/billingv1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestBuildBillingRouteGroupsPublishesCanonicalAndExactCompatibilitySurface(t *testing.T) {
	groups, err := buildBillingRouteGroups(
		new(billing.Handler),
		new(billingv1.Handler),
	)
	if err != nil {
		t.Fatal(err)
	}
	var routes []httpedge.Route
	for _, group := range groups {
		routes = append(routes, group...)
	}
	if len(groups) != 5 || len(routes) != 60 {
		t.Fatalf("Billing route groups=%d routes=%d", len(groups), len(routes))
	}
	if _, err := httpedge.NewRoutes(func(error) {}, routes...); err != nil {
		t.Fatalf("compile Billing routes: %v", err)
	}
}

func TestBuildBillingRouteGroupsFailsClosedOnMissingEdge(t *testing.T) {
	if _, err := buildBillingRouteGroups(nil, new(billingv1.Handler)); err == nil {
		t.Fatal("nil canonical Billing handler was accepted")
	}
	if _, err := buildBillingRouteGroups(new(billing.Handler), nil); err == nil {
		t.Fatal("nil Billing V1 compatibility handler was accepted")
	}
}
