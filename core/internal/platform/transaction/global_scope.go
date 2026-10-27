package transaction

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type globalScopeState struct {
	active            atomic.Bool
	actorID           uuid.UUID
	environmentID     uuid.UUID
	operationID       string
	mode              ScopeMode
	groupGeneration   int64
	billingGeneration int64
	transaction       *postgres.RuntimeTransaction
}

type GlobalScope struct {
	state *globalScopeState
}

func (scope GlobalScope) ActorID() uuid.UUID {
	if scope.state == nil || !scope.state.active.Load() {
		return uuid.Nil
	}
	return scope.state.actorID
}

func (scope GlobalScope) EnvironmentID() uuid.UUID {
	if scope.state == nil || !scope.state.active.Load() {
		return uuid.Nil
	}
	return scope.state.environmentID
}

func (scope GlobalScope) OperationID() string {
	if scope.state == nil || !scope.state.active.Load() {
		return ""
	}
	return scope.state.operationID
}

func (scope GlobalScope) Mode() ScopeMode {
	if scope.state == nil || !scope.state.active.Load() {
		return ""
	}
	return scope.state.mode
}
