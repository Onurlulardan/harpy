package postgres

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindProvider(
	scope platformtx.ProviderTenantScope,
	authority platformtx.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (Writer, error) {
	stamp, err := newAuthorityStamp(authority)
	if err != nil || !authority.OwnsProviderScope(scope) ||
		authority.Domain() != "billing" {
		return nil, ErrAuthorityMismatch
	}
	sealedJobs, sealedProducers, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("bind provider async release catalogs: %w", err)
	}
	jobs, err := jobsv1.NewCatalog(sealedJobs.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind provider async job catalog: %w", err)
	}
	producers, err := eventsv1.NewProducerCatalog(sealedProducers.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind provider async producer catalog: %w", err)
	}
	binder, err := platformpostgres.NewAuthorityBinder(
		storageOwner,
		func(database platformpostgres.BoundExecutor) Writer {
			return &repository{
				database: database, authority: stamp,
				operationID: scope.OperationID(), jobs: jobs, producers: producers,
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create provider async authority binder: %w", err)
	}
	writer, err := platformtx.BindProviderAuthority(scope, authority, binder)
	if err != nil {
		return nil, fmt.Errorf("bind provider async writer: %w", err)
	}
	return writer, nil
}
