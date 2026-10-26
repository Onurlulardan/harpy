package postgres

type AuthorityBinder[T any] struct {
	owner string
	build func(BoundExecutor) T
}

func NewAuthorityBinder[T any](
	owner string,
	build func(BoundExecutor) T,
) (AuthorityBinder[T], error) {
	if !schemaNamePattern.MatchString(owner) || build == nil {
		return AuthorityBinder[T]{}, ErrAuthorityBindingInvalid
	}
	return AuthorityBinder[T]{owner: owner, build: build}, nil
}

func (binder AuthorityBinder[T]) Owner() string { return binder.owner }

func BindAuthority[T any](
	transaction *RuntimeTransaction,
	binder AuthorityBinder[T],
) (T, error) {
	var zero T
	if transaction == nil || binder.build == nil ||
		!schemaNamePattern.MatchString(binder.owner) ||
		!transaction.bindings.usable(transaction) {
		return zero, ErrAuthorityBindingInvalid
	}
	executor := BoundExecutor{transaction: transaction, state: transaction.bindings}
	result := binder.build(executor)
	if rawBoundExecutorIdentity(any(result)) {
		return zero, ErrAuthorityBindingInvalid
	}
	return result, nil
}
