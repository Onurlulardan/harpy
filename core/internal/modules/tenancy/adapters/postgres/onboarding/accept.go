package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (storage *postgresStore) Accept(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
) (tenantapi.OnboardingOperation, error) {
	if !storage.usable(ctx) || !plan.IsValid() {
		return tenantapi.OnboardingOperation{}, errUnavailable
	}
	transition, err := scanOnboardingTransition(
		storage.queryRow(ctx, acceptSQL, onboardingAcceptArguments(plan)...),
	)
	if err != nil {
		return tenantapi.OnboardingOperation{}, fmt.Errorf("accept onboarding row: %w", err)
	}
	operation, err := tenantapi.NewOnboardingOperation(
		plan, transition.State, transition.Revision, transition.UpdatedAt,
	)
	if err != nil || transition.OperationID != plan.OperationID() {
		return tenantapi.OnboardingOperation{}, errUnavailable
	}
	return operation, nil
}
