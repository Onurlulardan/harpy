package billingv1

import (
	"context"
	"errors"
	"testing"
)

func TestAvailabilityPolicyUsesExactDeploymentMatrix(t *testing.T) {
	t.Parallel()
	saas, err := NewAvailabilityPolicy(DeploymentSaaS)
	if err != nil {
		t.Fatal(err)
	}
	onPrem, err := NewAvailabilityPolicy(DeploymentOnPrem)
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range Routes() {
		if err := saas.AdmitRoute(context.Background(), route); err != nil {
			t.Fatalf("SaaS route %s rejected: %v", route.ID, err)
		}
		err := onPrem.AdmitRoute(context.Background(), route)
		allowed := route.ID == "billing.route.subscription-read" ||
			route.ID == "billing.route.usage-read"
		if allowed && err != nil || !allowed && !errors.Is(err, ErrRouteUnavailable) {
			t.Fatalf("on-premise route %s allowed=%t error=%v", route.ID, allowed, err)
		}
	}
}

func TestUnavailableRouteUsesExactSaaSOnlyWire(t *testing.T) {
	t.Parallel()
	status, message, code := failureDetails("billing.route.public-plans", ErrRouteUnavailable)
	if status != 403 || code != "SAAS_ONLY_FEATURE" ||
		message != "This feature is only available in SaaS deployment mode" {
		t.Fatalf("availability wire = %d %q %q", status, message, code)
	}
	if _, err := NewAvailabilityPolicy("hosted"); err == nil {
		t.Fatal("unknown deployment mode was accepted")
	}
	events := []string{}
	availability := &orderAvailability{events: &events, err: ErrRouteUnavailable}
	retirement := &orderRetirement{events: &events}
	handler := newOrderHandler(t, availability, retirement, &events)
	if err := handler.Admit(context.Background()); err != nil {
		t.Fatal(err)
	}
	response := servePublicPlans(t, handler)
	wantBody := `{"error":"This feature is only available in SaaS deployment mode","code":"SAAS_ONLY_FEATURE"}`
	if response.Code != 403 || response.Body.String() != wantBody || retirement.hitCalls != 0 {
		t.Fatalf("unavailable response=%d body=%q hits=%d",
			response.Code, response.Body.String(), retirement.hitCalls)
	}
}
