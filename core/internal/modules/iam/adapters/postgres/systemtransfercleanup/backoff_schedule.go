package systemtransfercleanup

import (
	"context"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *Store) ScheduleSourceCleanupRetry(ctx context.Context,
	retry app.SourceCleanupRetry,
) (bool, error) {
	if !store.valid() || ctx == nil || ctx.Err() != nil ||
		retry.Target.Validate() != nil ||
		(retry.ErrorCode != app.SourceCleanupOutcomeUnknown &&
			retry.ErrorCode != app.SourceCleanupUnavailable) ||
		!validTime(retry.ObservedAt) || !validTime(retry.RetryAt) ||
		retry.RetryAt.Before(retry.ObservedAt) ||
		retry.RetryAt.Sub(retry.ObservedAt) > 15*time.Minute {
		return false, app.ErrSourceCleanupInvalid
	}
	var replay bool
	target := retry.Target
	err := store.runInTenant(ctx, target.TenantID,
		func(scoped context.Context, bound *tenantStore) error {
			schema, schemaErr := bound.schema(scoped)
			if schemaErr != nil {
				return schemaErr
			}
			return bound.database.QueryRow(scoped, retrySQL,
				schema, target.TenantID, target.SourceID, target.LeaseOwnerID,
				target.LeaseGeneration, retry.ErrorCode,
				retry.ObservedAt.UTC(), retry.RetryAt.UTC(),
			).Scan(&replay)
		})
	if err != nil {
		return false, mapError(err, store.notFound)
	}
	return replay, nil
}
