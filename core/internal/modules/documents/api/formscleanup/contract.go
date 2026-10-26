package formscleanup

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const (
	CatalogVersion        = "documents.forms-cleanup.v1"
	FormRecordSourceOwner = "FORMS"
	FormRecordSourceKind  = "form_record"
	FormRecordUsage       = "LIFECYCLE"
)

type State string

const (
	StateRequested State = "requested"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
)

type Request struct {
	OperationID, DefinitionID string
	Epoch                     int64
	RequestedAt               time.Time
}

type Advance struct{ OperationID string }

type Receipt struct {
	OperationID, DefinitionID            string
	Epoch                                int64
	CatalogVersion                       string
	LinkCount, ItemCount, VersionCount   int64
	CiphertextBytes                      int64
	LinkCatalogSHA256, ZeroBlockerSHA256 [32]byte
	ReceiptSHA256                        [32]byte
	IssuedAt                             time.Time
}

type Progress struct {
	OperationID, DefinitionID string
	Epoch                     int64
	State                     State
	Receipt                   *Receipt
}

type Owner interface {
	Request(context.Context, platformtx.TenantScope, Request) (Progress, error)
	Advance(context.Context, platformtx.TenantScope, Advance) (Progress, error)
}
