package postgres

import (
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Bind(pool *platformpostgres.Pool) (schemaexecutor.Persistence, error) {
	return platformpostgres.BindSchemaExecutorStore(
		pool,
		func(database *pgxpool.Pool, cleanup time.Duration) schemaexecutor.Persistence {
			return &schemaExecutorStore{pool: database, cleanupTimeout: cleanup}
		},
	)
}
