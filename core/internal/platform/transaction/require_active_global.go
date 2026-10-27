package transaction

// RequireActiveGlobal verifies the exact coordinator-owned global scope mode.
func RequireActiveGlobal(scope GlobalScope, mode ScopeMode) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrScopeClosed
	}
	if !mode.valid() || scope.state.mode != mode {
		return ErrScopeMode
	}
	return nil
}
