package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type CredentialDeliveryScope struct{ transaction *postgres.RuntimeTransaction }

func NewCredentialDeliveryScope(
	transaction *postgres.RuntimeTransaction,
) (CredentialDeliveryScope, error) {
	if transaction == nil {
		return CredentialDeliveryScope{}, errors.New("credential delivery transaction is nil")
	}
	return CredentialDeliveryScope{transaction: transaction}, nil
}

func BindCredentialDeliveryTenantSchema(
	scope CredentialDeliveryScope,
	ctx context.Context,
	schemaName string,
) error {
	if scope.transaction == nil {
		return errors.New("credential delivery transaction is nil")
	}
	return scope.transaction.BindActivationTenantSearchPath(ctx, schemaName)
}

func BindCredentialDelivery[T any](
	scope CredentialDeliveryScope,
	binder postgres.OwnerBinder[T],
) (T, error) {
	return postgres.BindOwner(scope.transaction, binder)
}
