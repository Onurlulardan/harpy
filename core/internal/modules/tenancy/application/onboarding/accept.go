package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) Accept(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
) (tenantapi.OnboardingOperation, error) {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil || !plan.IsValid() {
		return tenantapi.OnboardingOperation{}, tenantapi.ErrOnboardingInvalid
	}
	operation, err := coordinator.store.Accept(ctx, plan)
	if err != nil {
		return tenantapi.OnboardingOperation{}, fmt.Errorf("accept onboarding operation: %w", err)
	}
	if !operation.IsValid() || operation.Plan().StepPlanSHA256() != plan.StepPlanSHA256() {
		return tenantapi.OnboardingOperation{}, tenantapi.ErrOnboardingConflict
	}
	return operation, nil
}
