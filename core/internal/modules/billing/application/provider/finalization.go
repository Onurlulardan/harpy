package providerapp

import (
	"time"

	"github.com/google/uuid"
)

type FinalizeDecision string

const (
	FinalizePending          FinalizeDecision = "pending"
	FinalizeSucceeded        FinalizeDecision = "succeeded"
	FinalizeFailed           FinalizeDecision = "failed"
	FinalizeReconcile        FinalizeDecision = "reconcile"
	FinalizeReconcileUnknown FinalizeDecision = "reconcile_unknown"
	FinalizeReplay           FinalizeDecision = "replay"
)

func (decision FinalizeDecision) Valid() bool {
	switch decision {
	case FinalizePending, FinalizeSucceeded, FinalizeFailed,
		FinalizeReconcile, FinalizeReconcileUnknown, FinalizeReplay:
		return true
	default:
		return false
	}
}

func (decision FinalizeDecision) validState(state CommandState) bool {
	switch decision {
	case FinalizePending:
		return state == CommandQueued || state == CommandRunning ||
			state == CommandReconcileRequired
	case FinalizeSucceeded:
		return state == CommandSucceeded
	case FinalizeFailed:
		return state == CommandFailed
	case FinalizeReconcile, FinalizeReconcileUnknown:
		return state == CommandReconcileRequired
	case FinalizeReplay:
		return state == CommandSucceeded || state == CommandFailed
	default:
		return false
	}
}

type FinalizeCommandRequest struct {
	TenantID, JobID, AttemptID, CommandID uuid.UUID
	LeaseGeneration                       int64
	PayloadRevision                       int
	PayloadReference, PayloadSHA256       string
}

type FinalizeCommandMutation struct {
	FinalizeCommandRequest
	NextFinalizerJobID, ReconciliationJobID uuid.UUID
	ObservedAt                              time.Time
}

type Finalization struct {
	Decision                     FinalizeDecision
	CommandID, ActorID           uuid.UUID
	ActiveJobID, FinalizerJobID  uuid.UUID
	Provider, RequestSHA256      string
	State                        CommandState
	Revision                     int64
	OccurredAt                   time.Time
	ProviderReference, ErrorCode string
}

func (value Finalization) valid() bool {
	provider, err := CanonicalProviderName(value.Provider)
	return err == nil && provider == value.Provider &&
		value.Decision.Valid() && value.Decision.validState(value.State) &&
		value.CommandID != uuid.Nil && value.ActorID != uuid.Nil &&
		value.ActiveJobID != uuid.Nil && value.FinalizerJobID != uuid.Nil &&
		value.FinalizerJobID.Version() == uuid.Version(7) &&
		lowerSHA256Pattern.MatchString(value.RequestSHA256) &&
		value.State.Valid() && value.Revision > 0 &&
		!value.OccurredAt.IsZero() && value.OccurredAt.Location() == time.UTC &&
		value.OccurredAt.Nanosecond()%int(time.Microsecond) == 0
}
