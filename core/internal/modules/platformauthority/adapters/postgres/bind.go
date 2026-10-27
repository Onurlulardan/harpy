package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type transactionMode = platformpostgres.TransactionMode
type commandResult = platformpostgres.CommandResult
type databaseRow = platformpostgres.Row
type Pool = platformpostgres.Pool

const (
	transactionCommand       = platformpostgres.TransactionCommand
	transactionSnapshotQuery = platformpostgres.TransactionSnapshotQuery
)

type boundDatabase interface {
	Execute(context.Context, string, ...any) (commandResult, error)
	QueryRow(context.Context, string, ...any) databaseRow
}

type transactionRunner func(
	context.Context,
	transactionMode,
	func(boundDatabase) error,
) error

func NewStore(pool *platformpostgres.Pool, clock ports.Clock) (*Store, error) {
	if clock == nil {
		return nil, fmt.Errorf("platform authority postgres store is unavailable")
	}
	return platformpostgres.BindPlatformAuthorityStore(
		pool,
		func(database platformpostgres.PlatformAuthorityTransactions) (*Store, error) {
			runner := func(
				ctx context.Context,
				mode transactionMode,
				callback func(boundDatabase) error,
			) error {
				return database.Within(ctx, mode, func(bound platformpostgres.PlatformAuthorityDatabase) error {
					return callback(bound)
				})
			}
			return &Store{transactions: runner, clock: clock,
				rowNotFound: platformpostgres.RowNotFound}, nil
		},
	)
}
