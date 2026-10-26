package transaction

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type jobScopeState struct {
	active      atomic.Bool
	tenantID    uuid.UUID
	tenantSlug  string
	transaction *postgres.RuntimeTransaction
}

// JobTenantScope is a worker-only tenant transaction. It carries no actor and
// can bind only the exact owner selected by a generation-fenced job claim.
type JobTenantScope struct{ state *jobScopeState }

func (scope JobTenantScope) TenantID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.tenantID
}

func (scope JobTenantScope) TenantSlug() string {
	if !scope.valid() {
		return ""
	}
	return scope.state.tenantSlug
}

func (scope JobTenantScope) valid() bool {
	return scope.state != nil && scope.state.active.Load() &&
		scope.state.tenantID != uuid.Nil && scope.state.tenantSlug != "" &&
		scope.state.transaction != nil
}
