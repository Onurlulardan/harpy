package platformlicense

import (
	"context"
	"fmt"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type RuntimeDatabase interface {
	RuntimeOwnerDatabase()
}

func BindAPI(database RuntimeDatabase) (*Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("billing platform license database is unavailable")
	}
	return platformpostgres.BindBillingLicenseAPI(
		pool,
		func(database platformpostgres.BillingLicenseAPIDatabase) (*Store, error) {
			return NewStore(DatabaseAccess{
				QueryRow: func(
					ctx context.Context,
					statement string,
					arguments ...any,
				) Row {
					return database.QueryRow(ctx, statement, arguments...)
				},
			})
		},
	)
}
