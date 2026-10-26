package knowledgefile

import (
	"context"
	"io"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const (
	SourceOwner = "AI"
	SourceKind  = "knowledge_document"
	Usage       = "LIFECYCLE"
)

type StageRequest struct {
	OperationID, TenantID, ActorID ID
	AuthorizationVersion           int64
	Filename, MIME                 string
	Bytes                          int64
	SHA256                         [32]byte
	Content                        io.Reader
	ExpiresAt                      time.Time
}

type Staged struct {
	LeaseID, ItemID, VersionID ID
	Filename, MIME             string
	DetectedMIME, ScanRevision string
	Bytes                      int64
	SHA256                     [32]byte
	Clean                      bool
	ScannedAt, ExpiresAt       time.Time
}

type Adoption struct {
	LeaseID, ItemID, VersionID, ReferenceID ID
	SourceID                                ID
	IdempotencyKey                          string
}

type Release struct {
	ItemID, ReferenceID, SourceID ID
	IdempotencyKey                string
}

type OpenRequest struct {
	TenantID, ActorID, ItemID, VersionID ID
	AuthorizationVersion                 int64
}

type Opened struct {
	Content io.ReadCloser
	Bytes   int64
	SHA256  [32]byte
}

type ReconcileRequest struct {
	TenantID, ActorID    ID
	AuthorizationVersion int64
	Limit                uint16
}

type Reconciled struct {
	ExpiredItemIDs []ID
}

type Owner interface {
	Stage(context.Context, StageRequest) (Staged, error)
	Adopt(context.Context, platformtx.TenantScope, Adoption) error
	Release(context.Context, platformtx.TenantScope, Release) error
	Open(context.Context, OpenRequest) (Opened, error)
}

type Reconciler interface {
	ReconcileExpired(context.Context, ReconcileRequest) (Reconciled, error)
}

type FileOwner interface {
	Owner
	Reconciler
}
