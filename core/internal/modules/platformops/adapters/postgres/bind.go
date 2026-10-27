package postgres

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type transactionMode = platformpostgres.TransactionMode
type commandResult = platformpostgres.CommandResult
type databaseRow = platformpostgres.Row
type databaseRows = platformpostgres.Rows

const (
	transactionCommand       = platformpostgres.TransactionCommand
	transactionSnapshotQuery = platformpostgres.TransactionSnapshotQuery
)

type boundDatabase interface {
	Execute(context.Context, string, ...any) (commandResult, error)
	Query(context.Context, string, ...any) (databaseRows, error)
	QueryRow(context.Context, string, ...any) databaseRow
}

type transactionRunner func(
	context.Context, transactionMode, func(boundDatabase) error,
) error

type runtimeOwnerDatabase interface {
	RuntimeOwnerDatabase()
}

func NewStore(database runtimeOwnerDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("platform operations store binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsStore(
		pool,
		func(database platformpostgres.PlatformOpsTransactions) (*Store, error) {
			runner := func(ctx context.Context, mode transactionMode,
				callback func(boundDatabase) error,
			) error {
				return database.Within(ctx, mode,
					func(bound platformpostgres.PlatformOpsDatabase) error {
						return callback(bound)
					})
			}
			return &Store{transactions: runner,
				rowNotFound: platformpostgres.RowNotFound}, nil
		},
	)
}

func NewBackupStore(database runtimeOwnerDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("platform operations backup binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsBackupStore(
		pool,
		func(database platformpostgres.PlatformOpsTransactions) (*Store, error) {
			runner := func(ctx context.Context, mode transactionMode,
				callback func(boundDatabase) error,
			) error {
				return database.Within(ctx, mode,
					func(bound platformpostgres.PlatformOpsDatabase) error {
						return callback(bound)
					})
			}
			return &Store{transactions: runner,
				rowNotFound: platformpostgres.RowNotFound}, nil
		},
	)
}

func NewStoreForRestore(database runtimeOwnerDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("platform operations restore binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsRestoreStore(
		pool,
		func(database platformpostgres.PlatformOpsTransactions) (*Store, error) {
			runner := func(
				ctx context.Context,
				mode transactionMode,
				callback func(boundDatabase) error,
			) error {
				return database.Within(
					ctx, mode,
					func(bound platformpostgres.PlatformOpsDatabase) error {
						return callback(bound)
					},
				)
			}
			return &Store{
				transactions: runner,
				rowNotFound:  platformpostgres.RowNotFound,
			}, nil
		},
	)
}

func NewUnboundStoreForProfile(
	runner transactionRunner, rowNotFound func(error) bool,
) (*Store, error) {
	if runner == nil || rowNotFound == nil {
		return nil, fmt.Errorf("platform operations postgres profile is unavailable")
	}
	return &Store{transactions: runner, rowNotFound: rowNotFound}, nil
}

func BeginRecoverySnapshot(
	ctx context.Context,
	database any,
) (*platformpostgres.RecoverySnapshot, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("platform recovery snapshot database is unavailable")
	}
	return pool.BeginRecoverySnapshot(
		ctx,
		func(
			ctx context.Context,
			database platformpostgres.RecoverySnapshotDatabase,
		) (string, error) {
			var identifier string
			err := database.QueryRow(
				ctx,
				exportRecoverySnapshotSQL,
				"harpy-v2:core-database-backup-executor",
			).Scan(&identifier)
			return identifier, err
		},
	)
}
