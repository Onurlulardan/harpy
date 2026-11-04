package tenancy

import (
	"context"

	customerbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerbinding"
	provisioningpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerprovisioning"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/provisionlaunch"
	provisioningapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func NewTenantSchemaProvisioner(
	pool customerbindingpostgres.Pool,
) (provisioningapp.SchemaProvisioner, error) {
	return customerbindingpostgres.WithCustomerProvisioning(
		pool,
		func(database customerbindingpostgres.ProvisioningDatabase) (
			provisioningapp.SchemaProvisioner, error,
		) {
			return provisioningpostgres.NewSchemaStatus(func(
				ctx context.Context, statement string, arguments ...any,
			) provisioningpostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, provisionlaunch.Run)
		},
	)
}
