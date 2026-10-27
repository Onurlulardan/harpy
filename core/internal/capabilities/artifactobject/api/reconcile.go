package api

import (
	"context"
	"time"
)

type ReconciliationState string

const (
	ReconciliationSealed  ReconciliationState = "sealed"
	ReconciliationMissing ReconciliationState = "missing"
	ReconciliationUnknown ReconciliationState = "unknown"
)

type ReconcileExpected struct {
	Key              string
	CiphertextBytes  int64
	CiphertextSHA256 [32]byte
	HeaderSHA256     [32]byte
}

type ReconcileResult struct {
	State             ReconciliationState
	Identity          Identity
	ProviderRequestID string
	EvidenceSHA256    [32]byte
	ObjectModifiedAt  time.Time
	VerifiedAt        time.Time
}

type Reconciler interface {
	ReconcileExact(context.Context, ReconcileExpected) (ReconcileResult, error)
}
