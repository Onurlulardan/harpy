package systemtransfercleanup

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func (store *Store) AcquireSourceCleanup(ctx context.Context,
	claim app.SourceCleanupClaim,
) (app.SourceCleanupScan, error) {
	if !store.valid() || ctx == nil || ctx.Err() != nil ||
		claim.LeaseOwnerID.Version() != uuid.Version(7) || !validTime(claim.ObservedAt) {
		return app.SourceCleanupScan{}, app.ErrSourceCleanupInvalid
	}
	next, err := store.nextTenant(ctx, claim.AfterTenantID)
	if err != nil {
		return app.SourceCleanupScan{}, mapError(err, store.notFound)
	}
	if next.id == uuid.Nil {
		return app.SourceCleanupScan{}, nil
	}
	scan := app.SourceCleanupScan{ScannedTenantID: next.id}
	err = store.runInTenant(ctx, next.id,
		func(scoped context.Context, bound *tenantStore) error {
			target, found, claimErr := bound.claim(scoped, next, claim)
			scan.Target, scan.Found = target, found
			return claimErr
		})
	if err != nil {
		return app.SourceCleanupScan{}, mapError(err, store.notFound)
	}
	if scan.Validate() != nil {
		return app.SourceCleanupScan{}, app.ErrSourceCleanupInvalid
	}
	return scan, nil
}

func (store *tenantStore) claim(ctx context.Context, selected tenant,
	claim app.SourceCleanupClaim,
) (app.SourceCleanupTarget, bool, error) {
	schema, err := store.schema(ctx)
	if err != nil || schema != selected.schema {
		return app.SourceCleanupTarget{}, false,
			errors.Join(app.ErrSourceCleanupUnavailable, err)
	}
	var sourceID, actorID uuid.UUID
	var resource string
	target := app.SourceCleanupTarget{TenantID: selected.id,
		LeaseOwnerID: claim.LeaseOwnerID}
	err = store.database.QueryRow(ctx, claimSQL, schema, selected.id,
		claim.LeaseOwnerID, claim.ObservedAt.UTC()).Scan(
		&sourceID, &actorID, &resource, &target.ObjectKey, &target.ObjectVersion,
		&target.Attempt, &target.LeaseGeneration,
	)
	if store.notFound(err) {
		return app.SourceCleanupTarget{}, false, nil
	}
	parsed, valid := publicresource.ParseResource(resource)
	if err != nil || !valid {
		return app.SourceCleanupTarget{}, false, errors.Join(
			app.ErrSourceCleanupUnavailable, err,
		)
	}
	target.SourceID, target.ActorID, target.Resource = sourceID, actorID, parsed
	if target.Validate() != nil {
		return app.SourceCleanupTarget{}, false,
			app.ErrSourceCleanupInvalid
	}
	return target, true, nil
}

var _ app.SourceCleanupStore = (*Store)(nil)
