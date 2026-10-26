package lifecycle

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type OperationKind string
type OperationState string

const (
	OperationInstall     OperationKind = "INSTALL"
	OperationUpgrade     OperationKind = "UPGRADE"
	OperationRetry       OperationKind = "RETRY"
	OperationRestore     OperationKind = "RESTORE"
	OperationRetire      OperationKind = "RETIRE"
	OperationPurge       OperationKind = "PURGE"
	OperationDemoInstall OperationKind = "DEMO_INSTALL"
	OperationDemoRemove  OperationKind = "DEMO_REMOVE"

	OperationPlanned    OperationState = "PLANNED"
	OperationApplying   OperationState = "APPLYING"
	OperationWaiting    OperationState = "WAITING"
	OperationVerifying  OperationState = "VERIFYING"
	OperationActivating OperationState = "ACTIVATING"
	OperationCancelling OperationState = "CANCELLING"
	OperationSucceeded  OperationState = "SUCCEEDED"
	OperationBlocked    OperationState = "BLOCKED"
	OperationFailed     OperationState = "FAILED"
	OperationCancelled  OperationState = "CANCELLED"
)

type Operation struct {
	ID                 packagecontract.StableID
	Kind               OperationKind
	State              OperationState
	Plan               Plan
	RetryOf            *packagecontract.StableID
	StartedAt          time.Time
	DestructiveStarted *time.Time
	FailureCode        string
	Revision           uint64
}

func NewOperation(id packagecontract.StableID, kind OperationKind, plan Plan, at time.Time) (Operation, error) {
	if !id.Valid() || !validKind(kind) || !plan.Valid() || !canonicalTime(at) {
		return Operation{}, ErrInvalid
	}
	return Operation{ID: id, Kind: kind, State: OperationPlanned,
		Plan: plan.Clone(), StartedAt: at, Revision: 1}, nil
}

func (value Operation) Advance(next OperationState) (Operation, error) {
	if !operationTransition(value.State, next) {
		return Operation{}, ErrTransition
	}
	value.State, value.Revision = next, value.Revision+1
	return value, nil
}

func (value Operation) RequestCancel() (Operation, error) {
	if (value.Kind == OperationDemoInstall || value.Kind == OperationDemoRemove) &&
		value.State != OperationPlanned {
		return Operation{}, ErrCancellationLate
	}
	if value.State == OperationActivating || value.State == OperationSucceeded ||
		value.State == OperationFailed || value.State == OperationCancelled ||
		value.DestructiveStarted != nil {
		return Operation{}, ErrCancellationLate
	}
	if value.State == OperationPlanned {
		return value.Advance(OperationCancelled)
	}
	return value.Advance(OperationCancelling)
}

func (value Operation) StartDestructive(at time.Time) (Operation, error) {
	if value.Kind != OperationPurge || value.State != OperationApplying ||
		value.DestructiveStarted != nil || !canonicalTime(at) {
		return Operation{}, ErrTransition
	}
	value.DestructiveStarted, value.Revision = &at, value.Revision+1
	return value, nil
}

func validKind(kind OperationKind) bool {
	return kind == OperationInstall || kind == OperationUpgrade || kind == OperationRetry ||
		kind == OperationRestore || kind == OperationRetire || kind == OperationPurge ||
		kind == OperationDemoInstall || kind == OperationDemoRemove
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
