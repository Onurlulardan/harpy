package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type AuditIntent struct {
	ID                uuid.UUID
	Kind              string
	SubscriptionID    uuid.UUID
	ActorID           uuid.UUID
	SourceOperationID uuid.UUID
	BeforeRevision    int64
	AfterRevision     int64
	OccurredAt        time.Time
	Reason            string
}

func (intent AuditIntent) valid() bool {
	return validV7(intent.ID) && intent.Kind != "" && validID(intent.SubscriptionID) &&
		validV7(intent.SourceOperationID) && intent.BeforeRevision > 0 &&
		intent.AfterRevision == intent.BeforeRevision+1 &&
		validInstant(intent.OccurredAt) && len(intent.Reason) <= 512
}
