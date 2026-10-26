package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

const globalBillingOwner = "billing"

func BindGlobal[T any](
	scope GlobalScope,
	binder postgres.GlobalOwnerBinder[T],
) (T, error) {
	var zero T
	if scope.state == nil || !scope.state.active.Load() {
		return zero, ErrScopeClosed
	}
	if binder.Owner() != globalBillingOwner {
		return zero, postgres.ErrOwnerBindingUnauthorized
	}
	return postgres.BindGlobalOwner(scope.state.transaction, binder)
}
