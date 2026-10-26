package transaction

// RequireActive verifies that a capability is still inside its coordinator-owned
// callback and was opened with the exact mode required by the caller.
func RequireActive(scope TenantScope, mode ScopeMode) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrScopeClosed
	}
	if !mode.valid() || scope.state.mode != mode {
		return ErrScopeMode
	}
	return nil
}
