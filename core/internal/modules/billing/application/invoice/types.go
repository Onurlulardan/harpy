package invoice

import (
	"time"

	"github.com/google/uuid"

	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

type DraftSnapshot struct {
	InvoiceID    uuid.UUID
	Revision     int64
	Lines        []invoicedomain.Line
	Settled      money.Money
	PolicySHA256 [32]byte
	Finalized    bool
}

type FinalizeRequest struct {
	InvoiceID         uuid.UUID
	ActorID           uuid.UUID
	ExpectedRevision  int64
	SourceOperationID uuid.UUID
}

type Finalization struct {
	InvoiceID         uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
	ResultRevision    int64
	Totals            invoicedomain.Totals
	PolicySHA256      [32]byte
	FinalizedAt       time.Time
}

type FinalizeResult struct {
	Finalization Finalization
	Replay       bool
}

type FinalizeExecution struct {
	Request FinalizeRequest
	Now     time.Time
}

type FinalizationPlan struct {
	Result FinalizeResult
	Write  *Finalization
	Event  *FinalizedEventIntent
}
