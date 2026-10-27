package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) ReadOperation(
	ctx context.Context,
	scope platformtx.TenantScope,
	operationID uuid.UUID,
) (catalogdomain.Operation, error) {
	if operationID == uuid.Nil {
		return catalogdomain.Operation{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operationGetOperation, ActionRead)
	if err != nil {
		return catalogdomain.Operation{}, err
	}
	value, err := store.ReadOperation(ctx, operationID)
	if err != nil {
		return catalogdomain.Operation{}, ErrNotFound
	}
	if !value.ID.Valid() || !value.Plan.Valid() {
		return catalogdomain.Operation{}, ErrUnavailable
	}
	return value, nil
}
