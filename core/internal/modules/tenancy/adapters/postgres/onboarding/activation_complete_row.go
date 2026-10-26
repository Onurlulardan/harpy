package onboarding

import (
	"context"
	"fmt"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (storage *postgresStore) completeActivation(
	ctx context.Context,
	completion onboardingapp.ActivationCompletion,
) error {
	if !storage.usable(ctx) {
		return errUnavailable
	}
	if err := storage.execute(ctx, completeSQL, completion.OperationID,
		completion.ReceiptID, completion.ReceiptSHA256,
		completion.EvidenceSHA256, completion.ExpectedRevision); err != nil {
		return fmt.Errorf("complete activation row: %w", err)
	}
	return nil
}
