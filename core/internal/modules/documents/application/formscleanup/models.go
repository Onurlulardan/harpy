package formscleanup

import (
	"time"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Mutation struct {
	OperationID, DefinitionID domain.ID
	Epoch                     int64
	RequestedAt               time.Time
}

type Snapshot struct {
	Mutation
	State                                formscleanupapi.State
	LinkCount, ItemCount, VersionCount   int64
	CiphertextBytes                      int64
	LinkCatalogSHA256, ZeroBlockerSHA256 [32]byte
	ReceiptSHA256                        [32]byte
	CompletedAt                          time.Time
}

type Candidate struct {
	ItemID, JobID domain.ID
	PurgeRunAt    time.Time
}

type EnqueuePlan struct {
	OperationID, ItemID, JobID domain.ID
	RunAt                      time.Time
}

type Enqueued struct {
	OperationID, ItemID, JobID domain.ID
	RunAt, EnqueuedAt          time.Time
}
