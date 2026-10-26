package recoveryevidence

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type database = platformpostgres.RecoveryEvidenceDatabase
type transactionMode = platformpostgres.TransactionMode

const snapshotQuery = platformpostgres.TransactionSnapshotQuery

type transactionRunner func(
	context.Context,
	transactionMode,
	func(database) error,
) error

type Store struct{ transactions transactionRunner }

func Bind(database any) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("platform authority recovery evidence database is unavailable")
	}
	return platformpostgres.BindPlatformAuthorityRecoveryEvidenceStore(
		pool,
		func(transactions platformpostgres.RecoveryEvidenceTransactions) (*Store, error) {
			return &Store{transactions: transactions.Within}, nil
		},
	)
}

func (store *Store) within(
	ctx context.Context,
	callback func(database) error,
) error {
	if store == nil || store.transactions == nil || ctx == nil ||
		ctx.Err() != nil || callback == nil {
		return fmt.Errorf("platform authority recovery evidence store is unavailable")
	}
	return store.transactions(ctx, snapshotQuery, callback)
}
