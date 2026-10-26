package notifications

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func newPlatformNotificationBinder(catalogs platformasync.WriteCatalogs) (*postgres.Binder, error) {
	jobs, _, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read notification job catalog: %w", err)
	}
	binder, err := postgres.NewBinder(jobs,
		func(scope platformtx.TenantScope,
			authority platformtx.OwnerAuthority,
		) (postgres.EnqueueJob, error) {
			writer, bindErr := asyncpostgres.Bind(scope, authority, catalogs)
			if bindErr != nil {
				return nil, bindErr
			}
			return func(ctx context.Context, input jobsv1.Enqueue) (string, error) {
				result, enqueueErr := writer.EnqueueJob(ctx, input)
				return result.ID, enqueueErr
			}, nil
		})
	if err != nil {
		return nil, fmt.Errorf("build notification PostgreSQL binder: %w", err)
	}
	return binder, nil
}
