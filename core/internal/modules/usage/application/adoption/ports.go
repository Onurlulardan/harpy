package adoption

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Source interface {
	Inspect(context.Context) (Manifest, error)
	Page(context.Context, string, *uuid.UUID, int) (SourcePage, error)
}

type Target interface {
	Start(context.Context, StartCommand) (Run, bool, error)
	Claim(context.Context, ClaimCommand) (Run, error)
	ApplyPage(context.Context, ApplyPageCommand) (PageReceipt, error)
	Advance(context.Context, ApplyPageCommand) (Run, error)
	Reconcile(context.Context, LeaseCommand) (Run, error)
	Complete(context.Context, LeaseCommand) (Run, error)
}

type Clock interface{ Now() time.Time }

type StartCommand struct {
	OperationID, RunID uuid.UUID
	Manifest           Manifest
	TransformSHA256    [32]byte
	ObservedAt         time.Time
}

type ClaimCommand struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	ExpectedRevision               int64
	Manifest                       Manifest
	ObservedAt                     time.Time
}

type LeaseCommand struct {
	OperationID, RunID, LeaseOwner uuid.UUID
	LeaseGeneration                int64
	Manifest                       Manifest
	ObservedAt                     time.Time
}

type ApplyPageCommand struct {
	LeaseCommand
	PageNumber                     int64
	CursorBefore                   *uuid.UUID
	CursorAfter                    uuid.UUID
	Records                        []AdoptedRecord
	SourcePageSHA256, TargetSHA256 [32]byte
}

type PageReceipt struct {
	PageNumber int64
	RowCount   int
	State      string
}
