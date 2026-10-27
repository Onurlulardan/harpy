package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) Record(
	ctx context.Context,
	receipt tenantapi.OnboardingReceipt,
	expectedRevision int64,
) (tenantapi.OnboardingTransition, error) {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil ||
		!receipt.IsValid() || expectedRevision < 1 {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingInvalid
	}
	transition, err := coordinator.store.RecordReceipt(ctx, receipt, expectedRevision)
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("record onboarding receipt: %w", err)
	}
	if !transition.IsValid() || transition.OperationID != receipt.OperationID {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingConflict
	}
	return transition, nil
}
