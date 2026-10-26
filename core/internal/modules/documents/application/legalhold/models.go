package legalhold

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Source struct {
	Owner, CaseID string
}

type PlaceRequest struct {
	ItemID         domain.ID
	Source         Source
	EvidenceSHA256 [32]byte
	IdempotencyKey string
}

type ReleaseRequest struct {
	HoldID         domain.ID
	Source         Source
	IdempotencyKey string
}

type Mutation struct {
	Hold                          domain.LegalHold
	ActorID, OperationID, AuditID domain.ID
	Source                        Source
	IdempotencyKey                string
	RequestSHA256                 [32]byte
	OccurredAt                    time.Time
}
