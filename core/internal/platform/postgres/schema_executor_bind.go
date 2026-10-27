package postgres

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BindSchemaExecutorStore grants the owner adapter a profile-fenced store factory.
func BindSchemaExecutorStore(
	pool *Pool,
	factory func(*pgxpool.Pool, time.Duration) schemaexecutor.Persistence,
) (schemaexecutor.Persistence, error) {
	if pool == nil || pool.closed.Load() {
		return nil, ErrPoolClosed
	}
	if pool.profile != ProfileSchemaExecutor {
		return nil, fmt.Errorf("schema executor requires schema_executor postgres profile")
	}
	if factory == nil {
		return nil, fmt.Errorf("schema executor store factory is required")
	}
	return factory(pool.inner, pool.cleanupTimeout), nil
}
