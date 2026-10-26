package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type OutboxIntent struct {
	ID                uuid.UUID
	EventType         string
	EventVersion      int
	SubscriptionID    uuid.UUID
	AggregateRevision int64
	ActorID           uuid.UUID
	CausationID       uuid.UUID
	OccurredAt        time.Time
	ChangeKind        string
}

func (intent OutboxIntent) valid() bool {
	return validV7(intent.ID) && validOutboxEventType(intent.EventType) &&
		intent.EventVersion == 1 && validID(intent.SubscriptionID) &&
		intent.AggregateRevision > 0 && validV7(intent.CausationID) &&
		validInstant(intent.OccurredAt) && intent.ChangeKind != ""
}

func validOutboxEventType(value string) bool {
	return value == outboxSubscriptionChanged || value == outboxEntitlementsChanged
}
