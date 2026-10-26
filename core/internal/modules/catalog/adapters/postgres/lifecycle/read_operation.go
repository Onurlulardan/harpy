package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_operation.sql
var getOperationSQL string

func (storage *store) ReadOperation(
	ctx context.Context, operationID uuid.UUID,
) (catalogdomain.Operation, error) {
	if operationID == uuid.Nil {
		return catalogdomain.Operation{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return catalogdomain.Operation{}, unavailable()
	}
	return storage.loadOperation(ctx, operationID)
}

func (storage *store) loadOperation(ctx context.Context, operationID uuid.UUID) (catalogdomain.Operation, error) {
	value, _, err := scanOperation(storage.queryRow(ctx, getOperationSQL, operationID))
	if err != nil {
		if storage.notFound(err) {
			return catalogdomain.Operation{}, catalogapp.ErrNotFound
		}
		return catalogdomain.Operation{}, unavailable()
	}
	return value, nil
}
