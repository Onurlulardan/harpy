package transaction

// RequireOwner verifies that owner belongs to the exact active operation
// capability sealed into scope. It does not expose the operation owner set.
func RequireOwner(scope TenantScope, owner string) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrScopeClosed
	}
	operation := scope.state.operation
	if !authorityNamePattern.MatchString(owner) || !operation.valid() ||
		operation.mode != scope.state.mode || !operation.allowsOwner(owner) {
		return ErrScopeOwner
	}
	return nil
}
