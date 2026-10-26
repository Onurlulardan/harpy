package transaction

import "github.com/google/uuid"

// TenantSnapshotAuthority is an opaque projection of one exact owner fence
// captured by an active tenant SnapshotQuery scope.
type TenantSnapshotAuthority struct {
	state      *scopeState
	tenantID   uuid.UUID
	domain     string
	generation int64
}

func MintTenantSnapshotAuthority(
	scope TenantScope,
	owner string,
) (TenantSnapshotAuthority, error) {
	if err := RequireOwner(scope, owner); err != nil {
		return TenantSnapshotAuthority{}, err
	}
	if scope.Mode() != SnapshotQuery {
		return TenantSnapshotAuthority{}, ErrScopeOwner
	}
	for _, domain := range scope.state.operation.domains {
		if domain.name == owner && domain.generation > 0 {
			authority := TenantSnapshotAuthority{
				state: scope.state, tenantID: scope.state.tenantID,
				domain: owner, generation: domain.generation,
			}
			if authority.valid() {
				return authority, nil
			}
		}
	}
	return TenantSnapshotAuthority{}, ErrScopeOwner
}

func (authority TenantSnapshotAuthority) TenantID() uuid.UUID {
	if !authority.valid() {
		return uuid.Nil
	}
	return authority.tenantID
}

func (authority TenantSnapshotAuthority) Domain() string {
	if !authority.valid() {
		return ""
	}
	return authority.domain
}

func (authority TenantSnapshotAuthority) Generation() int64 {
	if !authority.valid() {
		return 0
	}
	return authority.generation
}

func (authority TenantSnapshotAuthority) Owns(scope TenantScope) bool {
	return authority.valid() && authority.state == scope.state
}

func (authority TenantSnapshotAuthority) valid() bool {
	return authority.state != nil && authority.state.active.Load() &&
		authority.state.mode == SnapshotQuery && authority.tenantID != uuid.Nil &&
		authority.domain != "" && authority.generation > 0 &&
		authority.tenantID == authority.state.tenantID
}
