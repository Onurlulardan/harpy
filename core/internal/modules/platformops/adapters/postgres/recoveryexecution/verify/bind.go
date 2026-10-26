package verify

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type databaseRow = platformpostgres.Row

type boundDatabase interface {
	QueryRow(context.Context, string, ...any) databaseRow
}

type transactionRunner func(
	context.Context,
	platformpostgres.TransactionMode,
	func(boundDatabase) error,
) error

type runtimeOwnerDatabase interface{ RuntimeOwnerDatabase() }

type Store struct{ transactions transactionRunner }

func Bind(database runtimeOwnerDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("recovery verify binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsBackupStore(
		pool,
		func(transactions platformpostgres.PlatformOpsTransactions) (*Store, error) {
			runner := func(
				ctx context.Context,
				mode platformpostgres.TransactionMode,
				callback func(boundDatabase) error,
			) error {
				return transactions.Within(
					ctx,
					mode,
					func(database platformpostgres.PlatformOpsDatabase) error {
						return callback(database)
					},
				)
			}
			return &Store{transactions: runner}, nil
		},
	)
}

func newUnbound(runner transactionRunner) *Store {
	return &Store{transactions: runner}
}

func (store *Store) withBound(
	ctx context.Context,
	callback func(boundDatabase) error,
) error {
	if store == nil || store.transactions == nil ||
		ctx == nil || ctx.Err() != nil || callback == nil {
		return fmt.Errorf("recovery verify store is unavailable")
	}
	return store.transactions(ctx, platformpostgres.TransactionCommand, callback)
}
