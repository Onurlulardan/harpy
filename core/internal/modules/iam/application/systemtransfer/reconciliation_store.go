package systemtransfer

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	ExportReconciliationMissingCode = "IAM_SYSTEM_EXPORT_ARTIFACT_MISSING"
	ExportReconciliationUnknownCode = "IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN"
	ExportReconciliationParkedCode  = "IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED"
)

type ExportReconciliationReceipt struct {
	EffectID, EffectReceiptID   uuid.UUID
	RequestSHA256, ResultSHA256 [32]byte
	ProviderRequestID           string
	State                       ExportArtifactReconciliationState
	ObservedAt                  time.Time
}

type ExportReconciliationUnknownRecord struct {
	Code          string
	FailureSHA256 [32]byte
	ObservedAt    time.Time
	RetryAt       time.Time
}

type ExportReconciliationStore interface {
	// A true claim starts or exactly replays the same read-only provider
	// observation. False is allowed only after an idempotent terminal resolution.
	AcquireReconciliation(
		context.Context,
		ExportReconciliationClaim,
	) (ExportReconciliationObservation, bool, error)
	PersistReconciliationSealed(
		context.Context, ExportReconciliationClaim,
		ExportReconciliationObservation, ExportArtifactReconciliation,
		ExportReconciliationReceipt,
	) error
	PersistReconciliationMissing(
		context.Context, ExportReconciliationClaim,
		ExportReconciliationObservation, ExportArtifactReconciliation,
		ExportReconciliationReceipt, string,
	) error
	ScheduleReconciliationRetry(
		context.Context, ExportReconciliationClaim,
		ExportReconciliationObservation, ExportReconciliationUnknownRecord,
	) error
	PersistReconciliationParked(
		context.Context, ExportReconciliationClaim,
		ExportReconciliationObservation, ExportReconciliationUnknownRecord,
	) error
}

type ExportReconciliationBackoff interface {
	RetryDelay(int) (time.Duration, error)
}
