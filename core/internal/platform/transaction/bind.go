package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

func Bind[T any](
	scope TenantScope,
	binder postgres.OwnerBinder[T],
) (T, error) {
	var zero T
	if err := RequireOwner(scope, binder.Owner()); err == ErrScopeClosed {
		return zero, ErrScopeClosed
	} else if err != nil {
		return zero, postgres.ErrOwnerBindingUnauthorized
	}
	return postgres.BindOwner(scope.state.transaction, binder)
}
