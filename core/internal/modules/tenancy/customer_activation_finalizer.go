package tenancy

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type CustomerActivationCallback func(
	context.Context,
	workerapp.Claim,
	OperatorActivationBinding,
	OperatorActivationResources,
) error

type CustomerActivationFinalizer struct {
	onboarding *onboardingapp.Coordinator
	activate   CustomerActivationCallback
}

func NewCustomerActivationFinalizer(
	onboarding *onboardingapp.Coordinator,
	activate CustomerActivationCallback,
) (*CustomerActivationFinalizer, error) {
	if onboarding == nil || activate == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &CustomerActivationFinalizer{
		onboarding: onboarding, activate: activate,
	}, nil
}

func (finalizer *CustomerActivationFinalizer) Finalize(
	ctx context.Context,
	claim workerapp.Claim,
	revision int64,
) error {
	if finalizer == nil || finalizer.onboarding == nil || finalizer.activate == nil ||
		!claim.Valid() || revision < claim.Revision {
		return tenantapi.ErrOnboardingInvalid
	}
	return finalizeActivation(finalizer.onboarding, ctx, OperatorFinalizationRequest{
		Plan: claim.Plan, LeaseTokenSHA256: claim.LeaseTokenSHA256,
		ExpectedRevision: revision,
	}, func(ctx context.Context, binding OperatorActivationBinding,
		resources OperatorActivationResources,
	) error {
		return finalizer.activate(ctx, claim, binding, resources)
	})
}

var _ workerapp.Finalizer = (*CustomerActivationFinalizer)(nil)
