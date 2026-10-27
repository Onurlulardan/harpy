package link

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Source struct {
	Owner, Kind   string
	ID, VersionID domain.ID
}

type RegisterRequest struct {
	ItemID         domain.ID
	Source         Source
	Usage          string
	ReferenceEpoch uint64
	IdempotencyKey string
}

type ReleaseRequest struct {
	LinkID         domain.ID
	Source         Source
	IdempotencyKey string
}

type Mutation struct {
	Link                          domain.Link
	ActorID, OperationID, AuditID domain.ID
	Source                        Source
	IdempotencyKey                string
	RequestSHA256                 [32]byte
	OccurredAt                    time.Time
}

type BlockerQuery struct {
	ItemID      domain.ID
	AfterLinkID domain.ID
	Limit       uint16
}

type BlockerPage struct {
	Items      []domain.Link
	NextLinkID domain.ID
}
