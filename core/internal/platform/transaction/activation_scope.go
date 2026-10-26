package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type ActivationScope struct{ transaction *postgres.RuntimeTransaction }

func NewActivationScope(transaction *postgres.RuntimeTransaction) (ActivationScope, error) {
	if transaction == nil {
		return ActivationScope{}, errors.New("activation transaction is nil")
	}
	return ActivationScope{transaction: transaction}, nil
}

func BindActivationTenantSchema(
	scope ActivationScope,
	ctx context.Context,
	schemaName string,
) error {
	if scope.transaction == nil {
		return errors.New("activation transaction is nil")
	}
	return scope.transaction.BindActivationTenantSearchPath(ctx, schemaName)
}

func BindActivation[T any](
	scope ActivationScope,
	binder postgres.OwnerBinder[T],
) (T, error) {
	return postgres.BindOwner(scope.transaction, binder)
}
