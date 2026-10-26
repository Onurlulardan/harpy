package postgres

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/jackc/pgx/v5/pgxpool"
)

const storageOwner = "platform_async"

type asyncWorkerPool = platformpostgres.Pool
type asyncWorkerDatabase = pgxpool.Pool
type cleanupDuration = time.Duration

func bindAsyncWorkerStore[T any](
	pool *asyncWorkerPool,
	factory func(*asyncWorkerDatabase, time.Duration) T,
) (T, error) {
	return platformpostgres.BindAsyncWorkerStore(pool, factory)
}

type boundDatabase interface {
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

func Bind(
	scope transaction.TenantScope,
	authority transaction.OwnerAuthority,
	catalogs platformasync.WriteCatalogs,
) (Writer, error) {
	stamp, err := newAuthorityStamp(authority)
	if err != nil {
		return nil, err
	}
	sealedJobs, sealedProducers, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("bind async release catalogs: %w", err)
	}
	jobs, err := jobsv1.NewCatalog(sealedJobs.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind async job catalog: %w", err)
	}
	producers, err := eventsv1.NewProducerCatalog(sealedProducers.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind async producer catalog: %w", err)
	}
	binder, err := platformpostgres.NewAuthorityBinder(
		storageOwner,
		func(database platformpostgres.BoundExecutor) Writer {
			return &repository{
				database: database, authority: stamp,
				principalID: scope.ActorID(), operationID: scope.OperationID(),
				jobs: jobs, producers: producers,
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create async authority binder: %w", err)
	}
	writer, err := transaction.BindAuthority(scope, authority, binder)
	if err != nil {
		return nil, fmt.Errorf("bind async writer: %w", err)
	}
	return writer, nil
}

func BindConsumer(
	scope transaction.TenantScope,
	authority transaction.OwnerAuthority,
	catalogs platformasync.ReleaseCatalogs,
	consumer string,
	consumerVersion int,
) (ConsumerStore, error) {
	jobs, events, err := catalogs.Snapshot()
	if err != nil {
		return nil, ErrConsumerBindingInvalid
	}
	return bindConsumerForCatalogs(
		scope, authority, jobs, events.Producers(), events.Consumers(),
		consumer, consumerVersion,
	)
}

func bindConsumerForCatalogs(
	scope transaction.TenantScope,
	authority transaction.OwnerAuthority,
	jobs jobsv1.Catalog,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	consumer string,
	consumerVersion int,
) (ConsumerStore, error) {
	stamp, err := newAuthorityStamp(authority)
	if err != nil {
		return nil, err
	}
	gapPolicy, err := resolveConsumerGapPolicy(
		stamp, jobs, producers, consumers, consumer, consumerVersion,
	)
	if err != nil {
		return nil, ErrConsumerBindingInvalid
	}
	binder, err := platformpostgres.NewAuthorityBinder(
		storageOwner,
		func(database platformpostgres.BoundExecutor) ConsumerStore {
			return &consumerRepository{
				database: database, authority: stamp, gapPolicy: gapPolicy,
				producers: producers, consumers: consumers,
				consumer: consumer, consumerVersion: consumerVersion,
				identity: &consumerIdentity{marker: 1},
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create async consumer binder: %w", err)
	}
	store, err := transaction.BindAuthority(scope, authority, binder)
	if err != nil {
		return nil, fmt.Errorf("bind async consumer: %w", err)
	}
	return store, nil
}

func bindWorkerRepository(
	pool *platformpostgres.Pool,
	coordinator *transaction.Coordinator,
) (*workerRepository, error) {
	catalogs, err := loadDatabaseJobCatalogs(coordinator)
	if err != nil {
		return nil, err
	}
	store, err := platformpostgres.BindAsyncWorkerStore(
		pool, func(database *pgxpool.Pool, cleanup time.Duration) *workerRepository {
			return &workerRepository{
				database: database, cleanup: cleanup,
				identity: &workerIdentity{marker: 1},
			}
		},
	)
	if err != nil {
		return nil, err
	}
	store.coordinator = coordinator
	store.databaseJobBinder, err = transaction.NewJobOwnerBinder(
		storageOwner,
		func(database platformpostgres.BoundExecutor) *boundDatabaseJobStore {
			return &boundDatabaseJobStore{
				database: database, jobs: catalogs.jobs, producers: catalogs.producers,
			}
		},
	)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func BindWorker(
	pool *platformpostgres.Pool,
	coordinator *transaction.Coordinator,
) (WorkerStore, error) {
	store, err := bindWorkerRepository(pool, coordinator)
	if err != nil {
		return nil, fmt.Errorf("bind async worker store: %w", err)
	}
	return store, nil
}

func BindPublisherStore(pool *platformpostgres.Pool) (PublisherStore, error) {
	store, err := bindWorkerRepository(pool, nil)
	if err != nil {
		return nil, fmt.Errorf("bind async publisher store: %w", err)
	}
	return store, nil
}

func BindObservability(
	pool *platformpostgres.Pool,
) (asyncobservability.Source, error) {
	source, err := bindAsyncWorkerStore(
		pool,
		func(database *asyncWorkerDatabase, _ cleanupDuration) asyncobservability.Source {
			return &observabilitySource{database: database}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("bind async observability source: %w", err)
	}
	return source, nil
}
