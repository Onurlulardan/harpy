package transaction

// RequireGlobalOwner admits only the fixed Billing owner in an active global scope.
func RequireGlobalOwner(scope GlobalScope, owner string) error {
	if RequireActiveGlobal(scope, SnapshotQuery) != nil {
		return ErrScopeOwner
	}
	if owner != globalBillingOwner {
		return ErrScopeOwner
	}
	return nil
}
