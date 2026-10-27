package dashboardexport

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDispatchCandidateChanged = errors.New("analytics dashboard export candidate changed")
	ErrDispatchQuotaWait        = errors.New("analytics dashboard export quota is unavailable")
	ErrDispatchLeaseStale       = errors.New("analytics dashboard export lease is stale")
	ErrSpoolCapacity            = errors.New("analytics dashboard export spool is unavailable")
	ErrWorkerCancelled          = errors.New("analytics dashboard export cancellation requested")
)

type RenderCandidate struct {
	TenantID, ExportID, JobID uuid.UUID
	TenantSchema              string
	AttemptGeneration         int64
}

type SpoolReservationIdentity struct {
	ExportID          uuid.UUID
	AttemptGeneration int64
	LeaseOwnerID      uuid.UUID
}

type EncryptedSpool interface {
	Identity() SpoolReservationIdentity
	OpenWriter() (io.WriteCloser, error)
	OpenReader() (io.ReadCloser, error)
	Discard() error
}

type RenderDispatchAttempt struct {
	Candidate             RenderCandidate
	Claim                 RenderClaim
	AttemptID             uuid.UUID
	ClaimedAt, LeaseUntil time.Time
}

type RenderExecutionState string

const (
	RenderExecutionReady          RenderExecutionState = "ready"
	RenderExecutionCancelled      RenderExecutionState = "cancelled"
	RenderExecutionRequeued       RenderExecutionState = "requeued"
	RenderExecutionDead           RenderExecutionState = "dead"
	RenderExecutionOutcomeUnknown RenderExecutionState = "outcome_unknown"
	RenderExecutionLeaseLost      RenderExecutionState = "lease_lost"
)

// RenderExecutionResult is returned only after the executor has durably
// recorded the matching job/effect outcome. CleanupError is populated when a
// known outcome permitted local spool disposal but disposal itself failed.
// Outcome-unknown and lease-lost executions intentionally retain the spool.
type RenderExecutionResult struct {
	State        RenderExecutionState
	CleanupError error
}

type RenderLease struct {
	LeaseUntil            time.Time
	CancellationRequested bool
}

type SpoolLeaseLookup interface {
	ActiveSpoolReservation(context.Context, SpoolReservationIdentity) (bool, error)
}

type SpoolAllocator interface {
	Reserve(SpoolReservationIdentity) (EncryptedSpool, error)
	Reap(context.Context, SpoolLeaseLookup, time.Time) error
}

type RenderClaimStore interface {
	SpoolLeaseLookup
	PeekRenderCandidate(context.Context) (RenderCandidate, bool, error)
	ClaimRenderCandidate(
		context.Context, RenderCandidate, SpoolReservationIdentity,
	) (RenderDispatchAttempt, error)
	CheckpointRender(
		context.Context, RenderDispatchAttempt, time.Time,
	) (RenderLease, error)
}

type RenderAttemptExecutor interface {
	ExecuteClaimedRender(
		context.Context, RenderDispatchAttempt, EncryptedSpool,
	) (RenderExecutionResult, error)
}

type DispatcherAdmission interface{ Accepting() bool }

const (
	renderDispatcherConcurrency = 4
	renderDispatcherIdlePoll    = 250 * time.Millisecond
	renderDispatcherReapEvery   = time.Hour
	renderDispatcherHeartbeat   = 20 * time.Second
)
