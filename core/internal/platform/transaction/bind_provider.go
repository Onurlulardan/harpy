package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

func BindProvider[T any](
	scope ProviderTenantScope,
	binder postgres.OwnerBinder[T],
) (T, error) {
	var zero T
	if !scope.valid() {
		return zero, ErrScopeClosed
	}
	if binder.Owner() != providerBindingOwner {
		return zero, postgres.ErrOwnerBindingUnauthorized
	}
	return postgres.BindOwner(scope.state.transaction, binder)
}
