package archive

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const StateReady = domain.ArchiveReady

type CreateRequest struct {
	ItemIDs          []domain.ID
	RetryOfArchiveID domain.ID
	IdempotencyKey   string
}

type GetRequest struct{ ArchiveID domain.ID }

type Snapshot struct {
	ID, TenantID, ActorID, JobID, RetryOfArchiveID domain.ID
	BillingReservationID                           domain.ID
	State                                          domain.ArchiveState
	AuthorizationRevision                          uint64
	EntryCount                                     int
	PlaintextBytes, ArtifactPlaintextBytes         int64
	CiphertextBytes                                int64
	RequestSHA256, EntriesSHA256                   [32]byte
	ArtifactSHA256, CiphertextSHA256               [32]byte
	ObjectKey, ObjectVersion, KeyID                string
	ObjectGeneration                               uint64
	WrappedDEK                                     []byte
	ReadyAt, ExpiresAt, CreatedAt, UpdatedAt       time.Time
}

type Entry struct {
	domain.ArchiveEntry
	TenantID     domain.ID
	Object       domain.ObjectReceipt
	OriginalName string
}

type CreateMutation struct {
	Snapshot
	OperationID, AuditID domain.ID
	RequestDocument      []byte
	IdempotencyKey       string
	OccurredAt           time.Time
}

type Plan struct {
	Snapshot
	Entries []Entry
}

type ArtifactReceipt struct {
	ObjectKey, ObjectVersion, KeyID   string
	ObjectGeneration                  uint64
	PlaintextBytes, CiphertextBytes   int64
	PlaintextSHA256, CiphertextSHA256 [32]byte
	WrappedDEK                        []byte
	WrittenAt                         time.Time
}
