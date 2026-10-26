package licenseauthority

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ComponentVersion struct {
	Component, Version string
}

type IssueRequest struct {
	OperationID       uuid.UUID
	Binding           Binding
	Nonce             [32]byte
	CurrentSequence   int64
	CurrentDigest     string
	ComponentVersions []ComponentVersion
}

type Source interface {
	IssueSnapshot(context.Context, IssueRequest) ([]byte, error)
}

type RefreshCommand struct {
	OperationID uuid.UUID
	Binding     Binding
}

type RefreshClaim struct {
	OperationID    uuid.UUID
	LeaseToken     uuid.UUID
	Binding        Binding
	Nonce          [32]byte
	Head           *AcceptedHead
	DatabaseNow    time.Time
	LeaseExpiresAt time.Time
}

type FailureKind string

const (
	FailureSource       FailureKind = "source_unavailable"
	FailureVerification FailureKind = "verification_failed"
	FailureRollback     FailureKind = "sequence_rollback"
	FailureConflict     FailureKind = "sequence_conflict"
)

type Store interface {
	AcquireRefreshLease(context.Context, RefreshCommand, []ComponentVersion) (RefreshClaim, error)
	PersistVerifiedSnapshot(context.Context, RefreshClaim, VerifiedSnapshot) error
	StoreRefreshFailure(context.Context, RefreshClaim, FailureKind) error
}

type ReconcileResult struct {
	Decision SequenceDecision
	Snapshot VerifiedSnapshot
}
