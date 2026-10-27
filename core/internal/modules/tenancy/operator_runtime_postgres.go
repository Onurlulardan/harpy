package tenancy

import (
	"context"
	"fmt"

	deploymentpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/deploymentpolicy"
	onboardingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/onboarding"
	operatorbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/operatorbinding"
	migrationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/tenantmigration"
	deploymentapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	migrationapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
	"github.com/google/uuid"
)

func BindOperatorRuntime(
	pool operatorbindingpostgres.Pool,
) (*OperatorRuntime, error) {
	return operatorbindingpostgres.Bind(pool, func(
		database operatorbindingpostgres.Database,
	) (*OperatorRuntime, error) {
		policyStore, err := deploymentpostgres.NewInitializer(func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) deploymentpostgres.Row {
			return database.QueryRow(ctx, statement, arguments...)
		})
		if err != nil {
			return nil, fmt.Errorf("bind operator deployment policy: %w", err)
		}
		policy, err := deploymentapp.NewInitializer(policyStore)
		if err != nil {
			return nil, fmt.Errorf("build operator deployment policy: %w", err)
		}
		operationStore, err := onboardingpostgres.NewStore(
			func(ctx context.Context, statement string, arguments ...any) onboardingpostgres.Row {
				return database.QueryRow(ctx, statement, arguments...)
			},
			database.Run,
		)
		if err != nil {
			return nil, fmt.Errorf("bind operator onboarding store: %w", err)
		}
		unit, err := onboardingpostgres.NewUnitOfWork(pool)
		if err != nil {
			return nil, fmt.Errorf("bind operator activation unit: %w", err)
		}
		onboarding, err := onboardingapp.New(operationStore, unit)
		if err != nil {
			return nil, fmt.Errorf("build operator onboarding: %w", err)
		}
		repository, err := migrationpostgres.NewRepository(func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) migrationpostgres.Row {
			return database.QueryRow(ctx, statement, arguments...)
		})
		if err != nil {
			return nil, fmt.Errorf("bind operator tenant migration: %w", err)
		}
		migration, err := migrationapp.NewService(repository, uuid.NewV7)
		if err != nil {
			return nil, fmt.Errorf("build operator tenant migration: %w", err)
		}
		return NewOperatorRuntime(policy, onboarding, migration)
	})
}
