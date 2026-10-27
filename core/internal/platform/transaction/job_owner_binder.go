package transaction

import "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

type JobBoundExecutor = postgres.BoundExecutor
type JobOwnerFactory[T any] func(JobBoundExecutor) T

type JobOwnerBinder[T any] struct {
	owner  string
	binder postgres.OwnerBinder[T]
}

func NewJobOwnerBinder[T any](
	owner string,
	build JobOwnerFactory[T],
) (JobOwnerBinder[T], error) {
	binder, err := postgres.NewOwnerBinder(owner, build)
	if err != nil {
		return JobOwnerBinder[T]{}, err
	}
	return JobOwnerBinder[T]{owner: owner, binder: binder}, nil
}

func (binder JobOwnerBinder[T]) Owner() string { return binder.owner }

func BindJob[T any](scope JobTenantScope, binder JobOwnerBinder[T]) (T, error) {
	var zero T
	if !scope.valid() || binder.owner == "" || binder.owner != binder.binder.Owner() {
		return zero, ErrScopeClosed
	}
	return postgres.BindOwner(scope.state.transaction, binder.binder)
}
