package recoverytarget

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type boundRecoveryDatabase struct {
	database platformpostgres.TenancyRecoveryDatabase
}

func (bound boundRecoveryDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) recoveryRow {
	return bound.database.QueryRow(ctx, statement, arguments...)
}

func Bind(database any) (*Reader, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("tenancy recovery database is unavailable")
	}
	return platformpostgres.BindTenancyRecoveryStore(
		pool,
		func(transactions platformpostgres.TenancyRecoveryTransactions) (*Reader, error) {
			within := func(
				ctx context.Context,
				callback func(recoveryDatabase) error,
			) error {
				return transactions.Within(
					ctx,
					platformpostgres.TransactionSnapshotQuery,
					func(database platformpostgres.TenancyRecoveryDatabase) error {
						return callback(boundRecoveryDatabase{database: database})
					},
				)
			}
			return &Reader{
				within:      within,
				rowNotFound: platformpostgres.RowNotFound,
			}, nil
		},
	)
}
