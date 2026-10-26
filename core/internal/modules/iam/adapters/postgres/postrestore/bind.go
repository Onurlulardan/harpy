package postrestore

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type RuntimePool interface {
	PostgresOwnerBindingPool() *platformpostgres.Pool
}

func NewStore(handle RuntimePool) (*Store, error) {
	if handle == nil {
		return nil, platformpostgres.ErrPoolClosed
	}
	pool := handle.PostgresOwnerBindingPool()
	return platformpostgres.BindIAMPostRestoreStore(pool,
		func(transactions platformpostgres.IAMPostRestoreTransactions) (*Store, error) {
			return &Store{
				withinGlobal: globalRunner(transactions),
				withinTenant: tenantRunner(transactions),
				rowNotFound:  platformpostgres.RowNotFound}, nil
		})
}

type databaseAdapter struct {
	database platformpostgres.IAMPostRestoreDatabase
}

type observerDatabaseAdapter struct {
	database platformpostgres.IAMPostRestoreObserverDatabase
}

func (adapter databaseAdapter) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) row {
	return adapter.database.QueryRow(ctx, statement, arguments...)
}

func (adapter observerDatabaseAdapter) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) row {
	return adapter.database.QueryRow(ctx, statement, arguments...)
}

func globalRunner(
	transactions platformpostgres.IAMPostRestoreTransactions,
) func(context.Context, transactionMode, func(database) error) error {
	return func(ctx context.Context, mode transactionMode, callback func(database) error) error {
		return transactions.WithinGlobal(ctx, postgresMode(mode),
			func(bound platformpostgres.IAMPostRestoreDatabase) error {
				return callback(databaseAdapter{database: bound})
			})
	}
}

func tenantRunner(
	transactions platformpostgres.IAMPostRestoreTransactions,
) func(context.Context, transactionMode, uuid.UUID, func(database) error) error {
	return func(
		ctx context.Context,
		mode transactionMode,
		tenantID uuid.UUID,
		callback func(database) error,
	) error {
		return transactions.WithinTenant(ctx, postgresMode(mode), tenantID,
			func(bound platformpostgres.IAMPostRestoreDatabase) error {
				return callback(databaseAdapter{database: bound})
			})
	}
}

func postgresMode(mode transactionMode) platformpostgres.TransactionMode {
	if mode == snapshotQuery {
		return platformpostgres.TransactionSnapshotQuery
	}
	return platformpostgres.TransactionCommand
}

func BindObserver(handle any) (*Observer, error) {
	pool, ok := handle.(*platformpostgres.Pool)
	if !ok || pool == nil {
		return nil, fmt.Errorf("IAM post-restore observer database is unavailable")
	}
	return &Observer{
		within: func(
			ctx context.Context,
			callback func(database) error,
		) error {
			return pool.WithinIAMPostRestoreObservation(
				ctx,
				func(bound platformpostgres.IAMPostRestoreObserverDatabase) error {
					return callback(observerDatabaseAdapter{database: bound})
				},
			)
		},
	}, nil
}
