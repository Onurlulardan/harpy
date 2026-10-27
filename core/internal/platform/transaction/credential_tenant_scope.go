package transaction

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type credentialScopeState struct {
	active      atomic.Bool
	target      VerifiedCredentialTarget
	transaction *postgres.RuntimeTransaction
}

// CredentialTenantScope opens only IAM's credential repository for one
// pre-authentication purpose. It cannot be converted to TenantScope.
type CredentialTenantScope struct{ state *credentialScopeState }

func (scope CredentialTenantScope) TenantID() uuid.UUID {
	if !scope.valid() {
		return uuid.Nil
	}
	return scope.state.target.tenantID
}

func (scope CredentialTenantScope) OperationID() string {
	if !scope.valid() {
		return ""
	}
	return scope.state.target.operationID
}

func (scope CredentialTenantScope) Purpose() CredentialPurpose {
	if !scope.valid() {
		return ""
	}
	return scope.state.target.purpose
}

func (scope CredentialTenantScope) LocatorKind() CredentialLocatorKind {
	if !scope.valid() {
		return ""
	}
	return scope.state.target.locatorKind
}

func (scope CredentialTenantScope) LocatorBindingSHA256() [32]byte {
	if !scope.valid() {
		return [32]byte{}
	}
	return scope.state.target.locatorBindingSHA256
}

func (scope CredentialTenantScope) SourceRevision() int64 {
	if !scope.valid() {
		return 0
	}
	return scope.state.target.sourceRevision
}

func (scope CredentialTenantScope) AuthorityGroupGeneration() int64 {
	if !scope.valid() {
		return 0
	}
	group, valid := scope.state.target.operation.singleGroup()
	if !valid {
		return 0
	}
	return group.generation
}

func (scope CredentialTenantScope) IAMGeneration() int64 {
	return scope.domainGeneration("iam")
}

func (scope CredentialTenantScope) SystemResourceGeneration() int64 {
	return scope.domainGeneration("system_resource_layout")
}

func (scope CredentialTenantScope) domainGeneration(domain string) int64 {
	if !scope.valid() {
		return 0
	}
	for _, expectation := range scope.state.target.operation.domains {
		if expectation.name == domain {
			return expectation.generation
		}
	}
	return 0
}

func (scope CredentialTenantScope) valid() bool {
	return scope.state != nil && scope.state.active.Load() && scope.state.target.valid()
}
