package systemtransfercleanup

import (
	"context"
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type tenant struct {
	id     uuid.UUID
	schema string
}

func (store *Store) nextTenant(ctx context.Context,
	after uuid.UUID,
) (tenant, error) {
	var afterValue any
	if after != uuid.Nil {
		afterValue = after
	}
	var next tenant
	err := store.database.QueryRow(ctx, nextTenantSQL, afterValue).Scan(
		&next.id, &next.schema,
	)
	if store.notFound(err) {
		return tenant{}, nil
	}
	if err != nil || next.id == uuid.Nil || !schemaPattern.MatchString(next.schema) {
		return tenant{}, errors.Join(app.ErrSourceCleanupUnavailable, err)
	}
	return next, nil
}
