package invoice

import (
	"time"

	"github.com/google/uuid"
)

const FinalizedEventType = "billing.invoice.finalized"

type FinalizedEventIntent struct {
	ID              uuid.UUID
	InvoiceID       uuid.UUID
	InvoiceRevision int64
	ActorID         uuid.UUID
	CausationID     uuid.UUID
	OccurredAt      time.Time
}

func (event FinalizedEventIntent) Valid() bool {
	return validV7(event.ID) && validV7(event.InvoiceID) &&
		event.InvoiceRevision > 0 && validID(event.ActorID) &&
		validV7(event.CausationID) && validInstant(event.OccurredAt)
}
