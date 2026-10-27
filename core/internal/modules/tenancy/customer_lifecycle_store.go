package tenancy

import (
	"context"

	customerbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerbinding"
	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerlifecycle"
	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

func NewCustomerLifecycleStore(
	pool customerbindingpostgres.Pool,
) (lifecycle.Store, error) {
	return customerbindingpostgres.WithCustomerProvisioning(pool,
		func(database customerbindingpostgres.ProvisioningDatabase) (lifecycle.Store, error) {
			return lifecyclepostgres.NewStore(func(ctx context.Context, statement string,
				arguments ...any,
			) lifecyclepostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, customerbindingpostgres.RowNotFound)
		})
}

func NewCustomerLifecycleWorkerStore(
	pool customerbindingpostgres.Pool,
) (lifecycle.WorkerStore, error) {
	return customerbindingpostgres.WithCustomerWorker(pool,
		func(database customerbindingpostgres.WorkerDatabase) (lifecycle.WorkerStore, error) {
			return lifecyclepostgres.NewStore(func(ctx context.Context, statement string,
				arguments ...any,
			) lifecyclepostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, customerbindingpostgres.RowNotFound)
		})
}
