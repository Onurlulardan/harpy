package schemaexecutor

import (
	"context"
	"time"
)

type Persistence interface {
	RecoverExpired(context.Context) (bool, error)
	Claim(context.Context, ClaimRequest) (ClaimRecord, bool, error)
	MarkRunning(context.Context, Lease) error
	Heartbeat(context.Context, HeartbeatRequest) error
	Apply(context.Context, ApplyRequest) error
	RecordFailure(context.Context, FailureRequest) error
}

type ClaimRequest struct {
	ExecutorID string
	LeaseID    string
}

type HeartbeatRequest struct {
	Lease Lease
}

type ApplyRequest struct {
	Lease Lease
	Plan  ValidatedPlan
}

type FailureState string

const (
	FailureRetryable FailureState = "retryable"
	FailureBlocked   FailureState = "blocked"
	FailureTerminal  FailureState = "failed"
)

type FailureRequest struct {
	Lease                Lease
	State                FailureState
	Code                 string
	RedactedDetailSHA256 string
	RetryDelay           time.Duration
}
