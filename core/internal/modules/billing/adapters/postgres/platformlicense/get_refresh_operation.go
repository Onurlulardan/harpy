package platformlicense

import (
	"context"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

func (store *Store) Get(
	ctx context.Context,
	environmentID uuid.UUID,
	operationID uuid.UUID,
) (platformlicenseapp.RefreshOperation, error) {
	if !store.usable(ctx) {
		return platformlicenseapp.RefreshOperation{}, platformlicenseapp.ErrUnavailable
	}
	value := platformlicenseapp.RefreshOperation{}
	err := store.database.QueryRow(
		ctx, readRefreshOperationSQL, environmentID, operationID,
	).Scan(
		&value.OperationID, &value.State, &value.SnapshotSequence,
		&value.SnapshotSHA256, &value.FailureCode, &value.CreatedAt,
		&value.CompletedAt,
	)
	if err != nil {
		return platformlicenseapp.RefreshOperation{}, classify(err)
	}
	value.CreatedAt = value.CreatedAt.UTC()
	if value.CompletedAt != nil {
		completed := value.CompletedAt.UTC()
		value.CompletedAt = &completed
	}
	return value, nil
}
