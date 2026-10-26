package api

import (
	"context"
	"errors"
	"testing"
	"time"

	billingv1 "github.com/Onurlulardan/harpy/core/internal/modules/billing/compat/billingv1"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func TestBillingAvailabilityUsesPersistedPolicy(t *testing.T) {
	t.Parallel()
	onPremise, err := billingV1Availability(billingPolicyFixture(
		t, tenancymodule.DeploymentModeOnPremise,
	))
	if err != nil {
		t.Fatal(err)
	}
	route := billingv1.Route{ID: "billing.route.public-plans"}
	if err := onPremise.AdmitRoute(context.Background(), route); !errors.Is(
		err, billingv1.ErrRouteUnavailable,
	) {
		t.Fatalf("on-premise route error = %v", err)
	}
	saas, err := billingV1Availability(billingPolicyFixture(
		t, tenancymodule.DeploymentModeSaaS,
	))
	if err != nil || saas.AdmitRoute(context.Background(), route) != nil {
		t.Fatalf("SaaS availability = %v, %v", saas, err)
	}
}

func TestBillingAvailabilityRejectsUnverifiedPolicy(t *testing.T) {
	t.Parallel()
	if _, err := billingV1Availability(tenancymodule.DeploymentPolicy{}); err == nil {
		t.Fatal("zero deployment policy was accepted")
	}
}

func billingPolicyFixture(
	t *testing.T,
	mode tenancymodule.DeploymentMode,
) tenancymodule.DeploymentPolicy {
	t.Helper()
	var maximum *int64
	if mode == tenancymodule.DeploymentModeOnPremise {
		zero := int64(0)
		maximum = &zero
	}
	policy, err := tenancymodule.NewDeploymentPolicy(
		"10000000-0000-4000-8000-000000000001",
		mode,
		maximum,
		time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
