package restoreexecution

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type transactionMode = platformpostgres.TransactionMode
type databaseRow = platformpostgres.Row

const transactionCommand = platformpostgres.TransactionCommand

type boundDatabase interface {
	QueryRow(context.Context, string, ...any) databaseRow
}

type transactionRunner func(
	context.Context, transactionMode, func(boundDatabase) error,
) error

type runtimeOwnerDatabase interface{ RuntimeOwnerDatabase() }

func Bind(database runtimeOwnerDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("restore execution binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsRestoreStore(
		pool,
		func(transactions platformpostgres.PlatformOpsTransactions) (*Store, error) {
			runner := func(
				ctx context.Context,
				mode transactionMode,
				callback func(boundDatabase) error,
			) error {
				return transactions.Within(
					ctx, mode,
					func(database platformpostgres.PlatformOpsDatabase) error {
						return callback(database)
					},
				)
			}
			return &Store{transactions: runner,
				rowNotFound: platformpostgres.RowNotFound}, nil
		},
	)
}

func NewUnbound(
	runner transactionRunner,
	rowNotFound func(error) bool,
) (*Store, error) {
	if runner == nil || rowNotFound == nil {
		return nil, fmt.Errorf("restore execution profile is unavailable")
	}
	return &Store{transactions: runner, rowNotFound: rowNotFound}, nil
}
