package postrestore

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/google/uuid"
)

func (store *Store) Binding(
	ctx context.Context, environmentID, restoreOperationID uuid.UUID,
) (application.Binding, error) {
	var result application.Binding
	var epoch []byte
	err := store.global(ctx, snapshotQuery, func(database database) error {
		return database.QueryRow(ctx, bindingSQL, environmentID, restoreOperationID).Scan(
			&result.TenantID, &result.OwnerUserID, &result.AnchorID,
			&result.NewDatabaseSystemID, &epoch, &result.TenantSlug)
	})
	if err == nil && len(epoch) == 32 {
		copy(result.SecurityEpoch[:], epoch)
	} else if err == nil {
		return application.Binding{}, application.ErrUnavailable
	}
	return result, err
}
