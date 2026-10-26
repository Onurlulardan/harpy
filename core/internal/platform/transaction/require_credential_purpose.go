package transaction

func RequireCredentialPurpose(
	scope CredentialTenantScope,
	purpose CredentialPurpose,
) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrScopeClosed
	}
	if !purpose.valid() || !scope.state.target.valid() ||
		scope.state.target.purpose != purpose {
		return ErrCredentialPurpose
	}
	return nil
}
