package transaction

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type providerScopeState struct {
	active      atomic.Bool
	lookup      ProviderLookup
	target      verifiedProviderTarget
	transaction *postgres.RuntimeTransaction
}

type ProviderTenantScope struct{ state *providerScopeState }

func (scope ProviderTenantScope) TenantID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.tenantID
}

func (scope ProviderTenantScope) MappingID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.mappingID
}

func (scope ProviderTenantScope) ProviderAccountID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.providerAccountID
}

func (scope ProviderTenantScope) SubscriptionID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.subscriptionID
}

func (scope ProviderTenantScope) InvoiceID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.invoiceID
}

func (scope ProviderTenantScope) OperationID() string {
	if !scope.valid() {
		return ""
	}
	return scope.state.lookup.OperationID()
}

func (scope ProviderTenantScope) valid() bool {
	return scope.state != nil && scope.state.active.Load() &&
		scope.state.lookup.valid() && scope.state.target.valid() &&
		scope.state.transaction != nil
}
