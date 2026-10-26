package adoption

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	RunPlanned     RunStatus = "planned"
	RunApplying    RunStatus = "applying"
	RunReconciling RunStatus = "reconciling"
	RunCompleted   RunStatus = "completed"
	RunFenced      RunStatus = "fenced"
)

type Run struct {
	OperationID, RunID, TenantID, CursorID, LeaseOwner uuid.UUID
	Kind                                               SourceKind
	SourceSchema, SourceRelation                       string
	SourceRelationOID                                  uint32
	TenantSlug, TenantName, SnapshotID                 string
	LocatorSHA256, ManifestSHA256, TransformSHA256     [32]byte
	ExpectedSourceSHA256, SourceSHA256, TargetSHA256   [32]byte
	ExpectedCount, CopiedCount, Revision               int64
	NextPage, LeaseGeneration                          int64
	Status                                             RunStatus
	LeaseExpiresAt, CreatedAt, UpdatedAt               time.Time
}

type StartCommand struct {
	OperationID, RunID uuid.UUID
	Locator            VerifiedLocator
	Manifest           SourceManifest
	TransformSHA256    [32]byte
	ObservedAt         time.Time
}

type ClaimCommand struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	ExpectedRevision               int64
	ObservedAt                     time.Time
}

type Record struct {
	SourceID     uuid.UUID       `json:"source_id"`
	SourceSHA256 string          `json:"source_sha256"`
	Target       json.RawMessage `json:"target"`
}

type ApplyPageCommand struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	Kind                           SourceKind
	LeaseGeneration, PageNumber    int64
	CursorBefore                   uuid.UUID
	Records                        []Record
	SourcePageSHA256               [32]byte
	ObservedAt                     time.Time
}

type PageReceipt struct {
	RunID, CursorBefore, CursorAfter   uuid.UUID
	PageNumber, RowCount               int64
	SourcePageSHA256, TargetPageSHA256 [32]byte
	Advanced                           bool
}

type LeaseCommand struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	LeaseGeneration                int64
	ObservedAt                     time.Time
}

type Target interface {
	AdoptionStart(context.Context, StartCommand) (Run, bool, error)
	AdoptionClaim(context.Context, ClaimCommand) (Run, error)
	AdoptionExecutePage(context.Context, ApplyPageCommand) (PageReceipt, error)
	AdoptionRecordAdvance(context.Context, ApplyPageCommand) (Run, error)
	AdoptionReconcile(context.Context, LeaseCommand) (Run, error)
	AdoptionComplete(context.Context, LeaseCommand) (Run, error)
}
