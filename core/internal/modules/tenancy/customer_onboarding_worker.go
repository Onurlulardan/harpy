package tenancy

import (
	"context"

	customerbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/customerbinding"
	onboardingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/onboarding"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func NewCustomerOnboardingCoordinator(
	pool customerbindingpostgres.Pool,
) (*onboardingapp.Coordinator, error) {
	store, err := customerbindingpostgres.WithCustomerWorker(pool,
		func(database customerbindingpostgres.WorkerDatabase) (
			onboardingapp.OperationStore, error,
		) {
			return onboardingpostgres.NewStore(func(
				ctx context.Context, statement string, arguments ...any,
			) onboardingpostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			}, func(ctx context.Context, statement string, arguments ...any) error {
				return database.Run(ctx, statement, arguments...)
			})
		})
	if err != nil {
		return nil, err
	}
	unit, err := onboardingpostgres.NewUnitOfWork(pool)
	if err != nil {
		return nil, err
	}
	return onboardingapp.New(store, unit)
}
