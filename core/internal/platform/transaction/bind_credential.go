package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

const credentialBindingOwner = "iam"

func BindCredential[T any](
	scope CredentialTenantScope,
	binder postgres.OwnerBinder[T],
) (T, error) {
	var zero T
	if !scope.valid() {
		return zero, ErrScopeClosed
	}
	if binder.Owner() != credentialBindingOwner {
		return zero, postgres.ErrOwnerBindingUnauthorized
	}
	return postgres.BindOwner(scope.state.transaction, binder)
}
