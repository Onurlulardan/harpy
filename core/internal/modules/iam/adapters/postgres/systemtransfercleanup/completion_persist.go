package systemtransfercleanup

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *Store) PersistSourceCleanupCompletion(ctx context.Context,
	completion app.SourceCleanupCompletion,
) (bool, error) {
	if !store.valid() || ctx == nil || ctx.Err() != nil ||
		completion.Target.Validate() != nil || completion.Receipt.Validate() != nil {
		return false, app.ErrSourceCleanupInvalid
	}
	var replay bool
	target := completion.Target
	err := store.runInTenant(ctx, target.TenantID,
		func(scoped context.Context, bound *tenantStore) error {
			schema, schemaErr := bound.schema(scoped)
			if schemaErr != nil {
				return schemaErr
			}
			return bound.database.QueryRow(scoped, completeSQL,
				schema, target.TenantID, target.SourceID, target.ActorID,
				string(target.Resource), target.ObjectKey, target.ObjectVersion,
				target.LeaseOwnerID, target.LeaseGeneration,
				completion.Receipt.ProviderRequestID,
				completion.Receipt.DeleteResponseSHA256[:],
				completion.Receipt.AbsenceProbeSHA256[:],
				completion.Receipt.VerifiedAbsentAt.UTC(),
			).Scan(&replay)
		})
	if err != nil {
		return false, mapError(err, store.notFound)
	}
	return replay, nil
}
