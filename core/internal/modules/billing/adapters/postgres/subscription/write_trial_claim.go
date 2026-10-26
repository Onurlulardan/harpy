package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) writeTrialClaim(
	ctx context.Context,
	authorityID uuid.UUID,
	claim *subscriptionapp.TrialClaim,
) error {
	if claim == nil {
		return nil
	}
	err := storage.expectID(
		ctx, insertTrialClaimSQL, claim.ID,
		authorityID, storage.tenantID, claim.ID, claim.SubscriptionID,
		claim.OfferingVersionID, claim.SourceOperationID, claim.ClaimedAt,
	)
	if err != nil {
		return fmt.Errorf("insert billing subscription trial claim: %w", err)
	}
	return nil
}
