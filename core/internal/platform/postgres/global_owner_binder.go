package postgres

const globalBillingOwner = "billing"

// GlobalOwnerBinder can only represent the ADR-0076 global Billing owner.
// Its distinct type prevents a tenant owner adapter from binding globally.
type GlobalOwnerBinder[T any] struct {
	owner string
	build func(BoundExecutor) T
}

func NewGlobalOwnerBinder[T any](
	owner string,
	build func(BoundExecutor) T,
) (GlobalOwnerBinder[T], error) {
	if owner != globalBillingOwner || build == nil {
		return GlobalOwnerBinder[T]{}, ErrOwnerBindingInvalid
	}
	return GlobalOwnerBinder[T]{owner: owner, build: build}, nil
}

func (binder GlobalOwnerBinder[T]) Owner() string { return binder.owner }

func BindGlobalOwner[T any](
	transaction *RuntimeTransaction,
	binder GlobalOwnerBinder[T],
) (T, error) {
	var zero T
	if transaction == nil || binder.owner != globalBillingOwner ||
		binder.build == nil || !transaction.bindings.usable(transaction) {
		return zero, ErrOwnerBindingInvalid
	}
	executor := BoundExecutor{transaction: transaction, state: transaction.bindings}
	result := binder.build(executor)
	if rawBoundExecutorIdentity(any(result)) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
