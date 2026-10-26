package environmentoperator

import (
	"context"
	"errors"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) finalize(
	ctx context.Context,
	request InitializeRequest,
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
) error {
	err := coordinator.finalizer.Finalize(ctx, FinalizationRequest{
		Manifest: request.Manifest, Plan: plan, AttemptID: claim.attemptID,
		LeaseTokenSHA256: claim.tokenSHA256,
		ExpectedRevision: claim.revision, LeaseUntil: claim.leaseUntil,
	})
	if err != nil {
		return classifyDependency(err)
	}
	operation, err := coordinator.onboarding.Accept(ctx, plan)
	if err != nil {
		if errors.Is(err, tenantapi.ErrOnboardingConflict) {
			return ErrFenced
		}
		return ErrResumable
	}
	if operation.State() != tenantapi.OnboardingSucceeded {
		return ErrFenced
	}
	return nil
}
