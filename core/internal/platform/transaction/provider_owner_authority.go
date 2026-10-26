package transaction

func MintProviderOwnerAuthority(
	scope ProviderTenantScope,
	owner string,
) (OwnerAuthority, error) {
	if !scope.valid() || owner != providerBindingOwner ||
		!scope.state.target.operation.allowsOwner(owner) {
		return OwnerAuthority{}, ErrScopeOwner
	}
	operation := scope.state.target.operation
	group, allowed := operation.groupForOwner(owner)
	if !allowed {
		return OwnerAuthority{}, ErrScopeOwner
	}
	for _, domain := range operation.domains {
		if domain.name == owner && domain.generation > 0 {
			return OwnerAuthority{
				providerState: scope.state, scopeKind: tenantAuthorityScopeKind,
				scopeID: scope.TenantID(), groupID: group.id,
				groupGen: group.generation, domain: owner,
				generation: domain.generation, origin: WriterOriginV2,
			}, nil
		}
	}
	return OwnerAuthority{}, ErrScopeOwner
}

func (authority OwnerAuthority) OwnsProviderScope(scope ProviderTenantScope) bool {
	return authority.valid() && authority.state == nil &&
		authority.globalState == nil && authority.providerState == scope.state &&
		authority.scopeKind == tenantAuthorityScopeKind
}
