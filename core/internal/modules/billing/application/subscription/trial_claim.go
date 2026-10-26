package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type TrialClaim struct {
	ID                uuid.UUID
	SubscriptionID    uuid.UUID
	OfferingVersionID uuid.UUID
	SourceOperationID uuid.UUID
	ClaimedAt         time.Time
}

func (claim TrialClaim) valid() bool {
	return validV7(claim.ID) && validID(claim.SubscriptionID) &&
		validID(claim.OfferingVersionID) && validV7(claim.SourceOperationID) &&
		validInstant(claim.ClaimedAt)
}
