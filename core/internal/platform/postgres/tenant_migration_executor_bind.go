package postgres

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

func BindTenantMigrationExecutorStore(
	pool *Pool,
	factory func(*pgxpool.Pool) tenantmigrationexecutor.Store,
) (tenantmigrationexecutor.Store, error) {
	if pool == nil || pool.closed.Load() {
		return nil, ErrPoolClosed
	}
	if pool.profile != ProfileTenantMigrationExecutor {
		return nil, fmt.Errorf(
			"tenant migration executor requires tenant_migration_executor postgres profile",
		)
	}
	if factory == nil {
		return nil, fmt.Errorf("tenant migration executor store factory is required")
	}
	return factory(pool.inner), nil
}
