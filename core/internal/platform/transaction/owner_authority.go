package transaction

import (
	"github.com/google/uuid"
)

const (
	tenantAuthorityScopeKind = "tenant"
	globalAuthorityScopeKind = "global"
)

type WriterOrigin string

const WriterOriginV2 WriterOrigin = "v2"

// OwnerAuthority is an opaque projection of an already-verified command scope.
type OwnerAuthority struct {
	state         *scopeState
	globalState   *globalScopeState
	providerState *providerScopeState
	scopeKind     string
	scopeID       uuid.UUID
	groupID       string
	groupGen      int64
	domain        string
	generation    int64
	origin        WriterOrigin
}

func MintOwnerAuthority(scope TenantScope, owner string) (OwnerAuthority, error) {
	if err := RequireOwner(scope, owner); err != nil {
		return OwnerAuthority{}, err
	}
	if scope.Mode() != Command {
		return OwnerAuthority{}, ErrScopeOwner
	}
	group, allowed := scope.state.operation.groupForOwner(owner)
	if !allowed {
		return OwnerAuthority{}, ErrScopeOwner
	}
	for _, domain := range scope.state.operation.domains {
		if domain.name == owner && domain.generation > 0 {
			return OwnerAuthority{
				state: scope.state, scopeKind: tenantAuthorityScopeKind,
				scopeID:  scope.state.tenantID,
				groupID:  group.id,
				groupGen: group.generation,
				domain:   owner, generation: domain.generation, origin: WriterOriginV2,
			}, nil
		}
	}
	return OwnerAuthority{}, ErrScopeOwner
}

func (authority OwnerAuthority) ScopeKind() string {
	if !authority.valid() {
		return ""
	}
	return authority.scopeKind
}

func (authority OwnerAuthority) ScopeID() uuid.UUID {
	if !authority.valid() {
		return uuid.Nil
	}
	return authority.scopeID
}

func (authority OwnerAuthority) GroupID() string {
	if !authority.valid() {
		return ""
	}
	return authority.groupID
}

func (authority OwnerAuthority) GroupGeneration() int64 {
	if !authority.valid() {
		return 0
	}
	return authority.groupGen
}

func (authority OwnerAuthority) Domain() string {
	if !authority.valid() {
		return ""
	}
	return authority.domain
}

func (authority OwnerAuthority) Generation() int64 {
	if !authority.valid() {
		return 0
	}
	return authority.generation
}

func (authority OwnerAuthority) Origin() WriterOrigin {
	if !authority.valid() {
		return ""
	}
	return authority.origin
}

func (authority OwnerAuthority) OwnsScope(scope TenantScope) bool {
	return authority.valid() && authority.scopeKind == tenantAuthorityScopeKind &&
		scope.state == authority.state && authority.providerState == nil
}

func (authority OwnerAuthority) OwnsGlobalScope(scope GlobalScope) bool {
	return authority.valid() && authority.scopeKind == globalAuthorityScopeKind &&
		scope.state == authority.globalState
}

func (authority OwnerAuthority) valid() bool {
	return authority.scopeActive() && authority.scopeID != uuid.Nil &&
		(authority.scopeKind == tenantAuthorityScopeKind ||
			authority.scopeKind == globalAuthorityScopeKind) && authority.groupID != "" &&
		authority.groupGen > 0 && authority.domain != "" &&
		authority.generation > 0 && authority.origin == WriterOriginV2
}

func (authority OwnerAuthority) scopeActive() bool {
	switch authority.scopeKind {
	case tenantAuthorityScopeKind:
		if authority.globalState != nil ||
			(authority.state == nil) == (authority.providerState == nil) {
			return false
		}
		if authority.state != nil {
			return authority.state.active.Load()
		}
		return authority.providerState.active.Load()
	case globalAuthorityScopeKind:
		return authority.globalState != nil && authority.state == nil &&
			authority.globalState.active.Load()
	default:
		return false
	}
}
