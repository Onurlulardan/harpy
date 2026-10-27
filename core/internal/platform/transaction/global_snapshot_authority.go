package transaction

import "github.com/google/uuid"

// GlobalSnapshotAuthority is the opaque generation fence captured by a live
// global Billing SnapshotQuery scope.
type GlobalSnapshotAuthority struct {
	state             *globalScopeState
	environmentID     uuid.UUID
	groupGeneration   int64
	billingGeneration int64
}

func MintGlobalSnapshotAuthority(
	scope GlobalScope,
	owner string,
) (GlobalSnapshotAuthority, error) {
	if RequireGlobalOwner(scope, owner) != nil {
		return GlobalSnapshotAuthority{}, ErrScopeOwner
	}
	authority := GlobalSnapshotAuthority{
		state: scope.state, environmentID: scope.state.environmentID,
		groupGeneration:   scope.state.groupGeneration,
		billingGeneration: scope.state.billingGeneration,
	}
	if !authority.valid() {
		return GlobalSnapshotAuthority{}, ErrScopeOwner
	}
	return authority, nil
}

func (authority GlobalSnapshotAuthority) EnvironmentID() uuid.UUID {
	if !authority.valid() {
		return uuid.Nil
	}
	return authority.environmentID
}

func (authority GlobalSnapshotAuthority) GroupGeneration() int64 {
	if !authority.valid() {
		return 0
	}
	return authority.groupGeneration
}

func (authority GlobalSnapshotAuthority) BillingGeneration() int64 {
	if !authority.valid() {
		return 0
	}
	return authority.billingGeneration
}

func (authority GlobalSnapshotAuthority) Owns(scope GlobalScope) bool {
	return authority.valid() && scope.state == authority.state
}

func (authority GlobalSnapshotAuthority) valid() bool {
	return authority.state != nil && authority.state.active.Load() &&
		authority.state.mode == SnapshotQuery && authority.environmentID != uuid.Nil &&
		authority.groupGeneration > 0 && authority.billingGeneration > 0 &&
		authority.environmentID == authority.state.environmentID &&
		authority.groupGeneration == authority.state.groupGeneration &&
		authority.billingGeneration == authority.state.billingGeneration
}
