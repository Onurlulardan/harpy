package recovery

import (
	"context"
	"fmt"

	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type Store struct {
	transactions    platformpostgres.DocumentsRecoveryTransactions
	environmentID   uuid.UUID
	tenantID        uuid.UUID
	tenantSetSHA256 [32]byte
}

type transactionMode = platformpostgres.TransactionMode
type recoveryDatabase = platformpostgres.DocumentsRecoveryDatabase

const (
	transactionCommand       = platformpostgres.TransactionCommand
	transactionSnapshotQuery = platformpostgres.TransactionSnapshotQuery
)

func Bind(
	database any,
	identity recoveryapp.Identity,
) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("documents recovery database is unavailable")
	}
	return platformpostgres.BindDocumentsRecoveryStore(
		pool,
		importSnapshot,
		func(transactions platformpostgres.DocumentsRecoveryTransactions) (*Store, error) {
			return &Store{transactions: transactions,
				environmentID: identity.EnvironmentID, tenantID: identity.TenantID,
				tenantSetSHA256: identity.TenantSetSHA256}, nil
		},
	)
}

func importSnapshot(
	ctx context.Context,
	database platformpostgres.DocumentsRecoveryDatabase,
	snapshotID string,
) error {
	_, err := database.Execute(ctx, importSnapshotSQL, snapshotID)
	return err
}

func (store *Store) with(
	ctx context.Context,
	mode transactionMode,
	callback func(recoveryDatabase) error,
) error {
	return store.transactions.WithinTenant(
		ctx, mode, store.environmentID, store.tenantID,
		store.tenantSetSHA256, callback,
	)
}

func (store *Store) withSnapshot(
	ctx context.Context,
	snapshotID string,
	callback func(recoveryDatabase) error,
) error {
	return store.transactions.WithinTenantSnapshot(
		ctx, store.environmentID, store.tenantID,
		store.tenantSetSHA256, snapshotID, callback,
	)
}

var _ recoveryapp.Store = (*Store)(nil)
