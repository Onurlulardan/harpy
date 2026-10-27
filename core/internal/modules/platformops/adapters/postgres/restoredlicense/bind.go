package restoredlicense

import (
	"context"
	"fmt"

	grantapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoredlicense"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type runtimeOwnerDatabase interface {
	RuntimeOwnerDatabase()
}

func NewStore(database runtimeOwnerDatabase) (grantapp.Store, error) {
	pool, ok := database.(*platformpostgres.Pool)
	if !ok {
		return nil, fmt.Errorf("restored license grant binding is unavailable")
	}
	return platformpostgres.BindPlatformOpsRestoredLicense(
		pool,
		func(bound platformpostgres.PlatformOpsRestoredLicenseDatabase) (
			grantapp.Store,
			error,
		) {
			return &store{
				queryRow: func(
					ctx context.Context,
					statement string,
					arguments ...any,
				) databaseRow {
					return bound.QueryRow(ctx, statement, arguments...)
				},
			}, nil
		},
	)
}
