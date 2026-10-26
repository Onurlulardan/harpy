package onboarding

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func TestAcceptPreservesImmutablePlan(t *testing.T) {
	plan := planFixture(t)
	coordinator, _, _ := coordinatorFixture(t, plan)
	operation, err := coordinator.Accept(context.Background(), plan)
	if err != nil || operation.Plan().StepPlanSHA256() != plan.StepPlanSHA256() {
		t.Fatalf("accepted operation = %#v, %v", operation, err)
	}
}

func TestAcceptFailsClosedOnPersistedPlanDrift(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, _ := coordinatorFixture(t, plan)
	store.operation = tenantapi.OnboardingOperation{}
	_, err := coordinator.Accept(context.Background(), plan)
	if !errors.Is(err, tenantapi.ErrOnboardingConflict) {
		t.Fatalf("plan drift error = %v", err)
	}
}
