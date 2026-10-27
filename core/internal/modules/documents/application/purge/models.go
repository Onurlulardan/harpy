package purge

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Version struct {
	ID                              domain.ID
	ObjectKey, ObjectVersion        string
	ObjectGeneration                uint64
	PlaintextBytes, CiphertextBytes int64
}

type Candidate struct {
	ItemID, TenantID, OwnerID domain.ID
	State                     domain.ItemState
	Versions                  []Version
}

type Deletion struct {
	Version
	DeletionID, JobID domain.ID
}

type Mutation struct {
	Candidate
	Deletions            []Deletion
	OperationID, AuditID domain.ID
	OccurredAt           time.Time
}

type Result struct {
	ItemID       domain.ID
	State        domain.ItemState
	PendingBlobs int
	PurgedAt     time.Time
}

func (result Result) Purged() bool { return result.State == domain.ItemPurged }

type DeletionPage struct {
	Items []Deletion
	More  bool
}
