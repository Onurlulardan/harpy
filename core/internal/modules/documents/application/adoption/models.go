package adoption

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const TransformVersion = "documents_adoption_v1"

type ObjectReceipt = domain.ObjectReceipt

type Authority struct {
	TenantID, RunID   domain.ID
	CutoverGeneration uint64
	ManifestSHA256    [32]byte
}

type Run struct {
	Authority
	State                                  domain.AdoptionState
	SourceCount, TargetCount               uint64
	SourceBytes, TargetBytes               uint64
	SourceMerkleSHA256, TargetMerkleSHA256 [32]byte
	PendingObjectCount, PendingJobCount    uint64
	FinalLSN, AppliedLSN                   string
	AcceptanceSHA256                       [32]byte
	CreatedAt, UpdatedAt                   time.Time
}

type CheckpointState string

const (
	CheckpointInventoried CheckpointState = "inventoried"
	CheckpointCopying     CheckpointState = "copying"
	CheckpointVerified    CheckpointState = "verified"
	CheckpointBlocked     CheckpointState = "blocked"
	CheckpointReleased    CheckpointState = "source_hold_released"
)

type Checkpoint struct {
	RunID, SourceItemID, SourceVersionID domain.ID
	TargetItemID, TargetVersionID        domain.ID
	SourceObjectKey, SourceRevision      string
	SourceBytes                          int64
	SourceSHA256, SourceMetadataSHA256   [32]byte
	SourceLSN                            string
	TargetObject                         domain.ObjectReceipt
	TargetMetadataSHA256                 [32]byte
	AppliedLSN, IssueCode                string
	State                                CheckpointState
	VerifiedAt, UpdatedAt                time.Time
}

type InventoryRequest struct {
	Authority
	Checkpoint
}

type CopyReceipt struct {
	TargetObject   domain.ObjectReceipt
	MetadataSHA256 [32]byte
	AppliedLSN     string
	VerifiedAt     time.Time
}

type ReconcileRequest struct {
	Authority
	FinalLSN, AppliedLSN string
	PendingJobCount      uint64
	OccurredAt           time.Time
}

type CompatibilityHit struct {
	RouteContract, ConsumerID, BuildID string
	HitCount                           uint64
	FirstHitAt, LastHitAt              time.Time
}

type ReverseRequest struct {
	ReconciliationID                   domain.ID
	SourceGeneration, TargetGeneration uint64
	SourceCount, TargetCount           uint64
	SourceSHA256, TargetSHA256         [32]byte
	RequestedBy                        domain.ID
	OccurredAt                         time.Time
}

type ReverseDecision struct {
	ReconciliationID, ApprovedBy domain.ID
	Accept                       bool
	OccurredAt                   time.Time
}
