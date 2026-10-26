package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) CurrentIdentity(
	ctx context.Context,
	expectedEnvironmentID string,
) (ports.EnvironmentIdentityRecord, error) {
	var record ports.EnvironmentIdentityRecord
	var mode string
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		identity := &record.Identity
		err := database.QueryRow(ctx, currentEnvironmentIdentitySQL,
			expectedEnvironmentID).Scan(
			&identity.EnvironmentID, &mode, &identity.PrimaryBindingID,
			&identity.TenantSetSHA256, &identity.OwnerBindingID,
			&identity.OwnerBindingGeneration, &identity.DatabaseLineageID,
			&record.DatabaseSystemID,
		)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.EnvironmentIdentityRecord{}, fmt.Errorf(
			"read current recovery identity: %w", err)
	}
	record.Identity.DeploymentMode = recoveryset.DeploymentMode(mode)
	if recoveryset.ValidateIdentity(record.Identity) != nil ||
		record.DatabaseSystemID == "" {
		return ports.EnvironmentIdentityRecord{}, ports.ErrConflict
	}
	return record, nil
}
