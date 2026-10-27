package postgres

type OwnerBinder[T any] struct {
	owner string
	build func(BoundExecutor) T
}

func NewOwnerBinder[T any](
	owner string,
	build func(BoundExecutor) T,
) (OwnerBinder[T], error) {
	if !schemaNamePattern.MatchString(owner) || build == nil {
		return OwnerBinder[T]{}, ErrOwnerBindingInvalid
	}
	return OwnerBinder[T]{owner: owner, build: build}, nil
}

func (binder OwnerBinder[T]) Owner() string { return binder.owner }

func BindOwner[T any](
	transaction *RuntimeTransaction,
	binder OwnerBinder[T],
) (T, error) {
	var zero T
	if transaction == nil || binder.build == nil ||
		!schemaNamePattern.MatchString(binder.owner) ||
		!transaction.bindings.usable(transaction) {
		return zero, ErrOwnerBindingInvalid
	}
	executor := BoundExecutor{transaction: transaction, state: transaction.bindings}
	result := binder.build(executor)
	if rawBoundExecutorIdentity(any(result)) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
