package transaction

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type scopeState struct {
	active      atomic.Bool
	actorID     uuid.UUID
	tenantID    uuid.UUID
	tenantSlug  string
	operation   VerifiedOperation
	mode        ScopeMode
	transaction *postgres.RuntimeTransaction
}

type TenantScope struct {
	state *scopeState
}

func (scope TenantScope) ActorID() uuid.UUID {
	if scope.state == nil {
		return uuid.Nil
	}
	return scope.state.actorID
}

func (scope TenantScope) TenantID() uuid.UUID {
	if scope.state == nil {
		return uuid.Nil
	}
	return scope.state.tenantID
}

func (scope TenantScope) TenantSlug() string {
	if scope.state == nil || !scope.state.active.Load() {
		return ""
	}
	return scope.state.tenantSlug
}

func (scope TenantScope) OperationID() string {
	if scope.state == nil {
		return ""
	}
	return scope.state.operation.operationID
}

func (scope TenantScope) Mode() ScopeMode {
	if scope.state == nil {
		return ""
	}
	return scope.state.mode
}
