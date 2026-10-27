package tenancy

import (
	"context"

	customerbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerbinding"
	workerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerworker"
	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
)

func NewCustomerWorkerStore(
	pool customerbindingpostgres.Pool,
) (workerapp.ClaimStore, error) {
	return customerbindingpostgres.WithCustomerWorker(pool,
		func(database customerbindingpostgres.WorkerDatabase) (
			workerapp.ClaimStore, error,
		) {
			return workerpostgres.NewStore(func(
				ctx context.Context, statement string, arguments ...any,
			) workerpostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, customerbindingpostgres.RowNotFound)
		})
}
