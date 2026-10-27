package upload

import (
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const UploadKeep = time.Hour

type Snapshot struct {
	ID, TenantID, ActorID                domain.ID
	DocumentID, CandidateVersionID       domain.ID
	ExpectedCurrentVersionID, FolderID   domain.ID
	BillingReservationID                 domain.ID
	RequestSHA256, ExpectedContentSHA256 [32]byte
	ExpectedBytes, CiphertextUpperBound  int64
	OriginalName, DeclaredMIME           string
	Description                          string
	Tags                                 []string
	AttemptGeneration                    uint64
	State                                domain.UploadState
	CleanupRequired                      bool
	StableErrorCode                      string
	ReadyStatus                          int
	ReadyResponse                        []byte
	ExpiresAt, CreatedAt, UpdatedAt      time.Time
	CancelRequestedAt                    time.Time
}

type CreateRequest struct {
	FolderID, DocumentID, ExpectedCurrentVersionID domain.ID
	OriginalName, DeclaredMIME, IdempotencyKey     string
	Description                                    string
	Tags                                           []string
	ExpectedContentSHA256                          [32]byte
	ExpectedBytes                                  int64
}

type CreateMutation struct {
	Snapshot
	OperationID, AuditID domain.ID
	IdempotencyKey       string
}

type GetRequest struct{ UploadID domain.ID }

type PutRequest struct {
	UploadID, LeaseOwnerID domain.ID
	Content                io.Reader
}

type AttemptPlan struct {
	UploadID, TenantID, ActorID           domain.ID
	DocumentID, CandidateVersionID        domain.ID
	ExpectedCurrentVersionID              domain.ID
	BillingReservationID, EffectReceiptID domain.ID
	Generation                            uint64
	ObjectKey, OriginalName, DeclaredMIME string
	ExpectedContentSHA256                 [32]byte
	ExpectedBytes                         int64
	DeclaredKind                          domain.FileKind
}

type EffectReceipt struct {
	Plan                       AttemptPlan
	Object                     domain.ObjectReceipt
	DetectedKind               domain.FileKind
	DetectedMIME, ScanRevision string
	ScannedAt                  time.Time
}

type EffectFailure struct {
	Plan       AttemptPlan
	Kind       EffectFailureKind
	StableCode string
	Object     *domain.ObjectReceipt
	OccurredAt time.Time
}

type FinalizeDecision struct {
	Ready, Conflict, Cancelled bool
	CurrentVersionID           domain.ID
}

type ReadyResult struct {
	Upload                Snapshot
	DocumentID, VersionID domain.ID
	Status                int
	Response              []byte
}

type CancelRequest struct{ UploadID domain.ID }

type CancelResult struct {
	Snapshot
	ReleaseReservation bool
	ReleaseReceiptID   domain.ID
}
