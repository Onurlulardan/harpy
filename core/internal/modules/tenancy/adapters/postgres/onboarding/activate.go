package onboarding

import (
	"context"
	"fmt"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (transaction *activationTransaction) Activate(
	ctx context.Context,
	request onboardingapp.ActivationRequest,
	completion onboardingapp.ActivationCompletion,
	callback onboardingapp.ActivationTransactionCallback,
) error {
	if !transaction.usable() || ctx == nil || ctx.Err() != nil || callback == nil {
		return errUnavailable
	}
	claim, err := transaction.store.claimActivation(ctx, request)
	if err != nil {
		return fmt.Errorf("claim activation: %w", err)
	}
	resources := onboardingapp.ActivationResources{
		BindTenantSchema: transaction.bindTenantSchema,
		BindOwner:        transaction.bindOwner,
	}
	if err := callback(claim, resources); err != nil {
		return err
	}
	completion.ExpectedRevision = claim.Revision
	return transaction.store.completeActivation(ctx, completion)
}
