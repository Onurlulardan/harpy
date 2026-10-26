package systemtransfercleanup

import (
	"context"
	"fmt"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const storageOwner = "iam"

func Bind(pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
) (*Store, error) {
	if coordinator == nil {
		return nil, fmt.Errorf("bind IAM import cleanup: coordinator is unavailable")
	}
	store, err := platformpostgres.BindAsyncWorkerExecutor(pool,
		func(database platformpostgres.AsyncWorkerExecutor) *Store {
			return &Store{database: databaseFuncs{queryRow: database.QueryRow},
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, fmt.Errorf("bind IAM import cleanup inventory: %w", err)
	}
	owner, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformtx.JobBoundExecutor) *tenantStore {
			return &tenantStore{database: databaseFuncs{queryRow: database.QueryRow},
				currentSchema: database.CurrentSchema,
				notFound:      platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, fmt.Errorf("bind IAM import cleanup owner: %w", err)
	}
	store.inTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *tenantStore) error,
	) error {
		return coordinator.InJobTenant(ctx, tenantID,
			func(scoped context.Context, scope platformtx.JobTenantScope) error {
				bound, bindErr := platformtx.BindJob(scope, owner)
				if bindErr != nil || bound == nil || platformtx.NetworkAllowed(scoped) ||
					!platformtx.ContextOwnsJobScope(scoped, scope) {
					return app.ErrSourceCleanupUnavailable
				}
				return execute(scoped, bound)
			})
	}
	return store, nil
}

type databaseFuncs struct {
	queryRow func(context.Context, string, ...any) platformpostgres.Row
}

func (database databaseFuncs) QueryRow(ctx context.Context,
	statement string, arguments ...any,
) row {
	return database.queryRow(ctx, statement, arguments...)
}
