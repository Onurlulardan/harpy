package tenancy

import (
	"context"

	customerbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerbinding"
	provisioningpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerprovisioning"
	provisioningapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func NewCustomerProvisioningStore(
	pool customerbindingpostgres.Pool,
) (provisioningapp.OperationStore, error) {
	return customerbindingpostgres.WithCustomerProvisioning(
		pool,
		func(database customerbindingpostgres.ProvisioningDatabase) (
			provisioningapp.OperationStore, error,
		) {
			return provisioningpostgres.NewStore(func(
				ctx context.Context, statement string, arguments ...any,
			) provisioningpostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, customerbindingpostgres.RowNotFound)
		},
	)
}
