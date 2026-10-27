package systemtransfer

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ExportClaimResult struct {
	State  ExportOwnerState
	Replay bool
}

type ExportProgress struct {
	Rows, PlaintextBytes int64
	CheckedAt            time.Time
}

func (progress ExportProgress) Validate() error {
	if progress.Rows < 0 || progress.Rows > 1_000_000 ||
		progress.PlaintextBytes < 0 || progress.PlaintextBytes > 2<<30 ||
		!exportArtifactTime(progress.CheckedAt) {
		return ErrExportWorkerInvalid
	}
	return nil
}

type ExportCompletion struct {
	Attempt  ExportAttempt
	Progress ExportProgress
	Artifact ExportArtifactMetadata
}

type ExportFinalState string

const (
	ExportFinalFailed    ExportFinalState = "failed"
	ExportFinalCancelled ExportFinalState = "cancelled"
)

type ExportFailure struct {
	Attempt ExportAttempt
	State   ExportFinalState
	Code    string
	At      time.Time
}

type ExportWorkerStore interface {
	AcquireOutbound(context.Context, ExportAttempt) (ExportClaimResult, error)
	CheckpointExport(context.Context, ExportAttempt, ExportProgress) error
	BeginArtifactEffect(context.Context, ExportArtifactEffectStart) (bool, error)
	PersistOutboundSuccess(context.Context, ExportCompletion) (bool, error)
	PersistOutcomeUnknown(context.Context, ExportOutcomeUnknownRecord) error
	PersistOutboundFailure(context.Context, ExportFailure) error
}

type ExportWorkerStoreBinder interface {
	BindExportWorker(context.Context, platformtx.TenantScope) (ExportWorkerStore, error)
}
