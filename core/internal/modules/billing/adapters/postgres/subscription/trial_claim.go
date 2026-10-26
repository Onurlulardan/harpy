package subscription

import (
	"context"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) findTrialClaim(
	ctx context.Context,
	authorityID, offeringID uuid.UUID,
) (*subscriptionapp.TrialClaim, error) {
	var claim subscriptionapp.TrialClaim
	var claimedAt time.Time
	err := storage.queryRow(
		ctx, findTrialClaimSQL, authorityID, storage.tenantID, offeringID,
	).Scan(
		&claim.ID, &claim.SubscriptionID, &claim.OfferingVersionID,
		&claim.SourceOperationID, &claimedAt,
	)
	if err != nil {
		if storage.notFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find billing subscription trial claim: %w", err)
	}
	claim.ClaimedAt = claimedAt.UTC()
	return &claim, nil
}
