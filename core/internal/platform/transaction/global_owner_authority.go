package transaction

func MintGlobalOwnerAuthority(
	scope GlobalScope,
	owner string,
) (OwnerAuthority, error) {
	if scope.state == nil || !scope.state.active.Load() {
		return OwnerAuthority{}, ErrScopeClosed
	}
	if scope.state.mode != Command || owner != globalBillingOwner {
		return OwnerAuthority{}, ErrScopeOwner
	}
	return OwnerAuthority{
		globalState: scope.state, scopeKind: globalAuthorityScopeKind,
		scopeID: scope.state.environmentID, groupID: globalBillingOwner,
		groupGen: scope.state.groupGeneration, domain: globalBillingOwner,
		generation: scope.state.billingGeneration, origin: WriterOriginV2,
	}, nil
}
