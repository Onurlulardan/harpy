package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindProcessor(
	scope platformtx.JobTenantScope,
	nextID subscriptionapp.IDGenerator,
	clock subscriptionapp.Clock,
	jobs JobWriter,
	update SubscriptionUpdater,
) (providerapp.WebhookProcessorStore, error) {
	if nextID == nil || clock == nil || jobs == nil || update == nil {
		return nil, providerapp.ErrConfiguration
	}
	binder, err := platformtx.NewJobOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *processorStore {
			return &processorStore{
				tenantID: scope.TenantID(), database: boundAccess(database, scope),
				nextID: nextID, clock: clock, jobs: jobs, update: update,
			}
		},
	)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	bound, err := platformtx.BindJob(scope, binder)
	if err != nil || bound == nil {
		return nil, providerapp.ErrConfiguration
	}
	return bound, nil
}

func boundAccess(
	database platformpostgres.BoundExecutor,
	scope platformtx.JobTenantScope,
) databaseAccess {
	return databaseAccess{
		queryRow: func(ctx context.Context, statement string, arguments ...any) databaseRow {
			return database.QueryRow(ctx, statement, arguments...)
		},
		execute: func(ctx context.Context, statement string, arguments ...any) error {
			_, err := database.Execute(ctx, statement, arguments...)
			return err
		},
		allowed: func(ctx context.Context) bool {
			return platformtx.ContextOwnsJobScope(ctx, scope) &&
				!platformtx.NetworkAllowed(ctx)
		},
		notFound: platformpostgres.RowNotFound,
	}
}
