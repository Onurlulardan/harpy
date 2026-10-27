package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func configurePool(
	poolConfig *pgxpool.Config,
	settings config.Postgres,
	contract profileContract,
) {
	poolConfig.ConnConfig.ConnectTimeout = settings.ConnectTimeout()
	poolConfig.ConnConfig.RuntimeParams["role"] = contract.currentUser
	poolConfig.ConnConfig.RuntimeParams["search_path"] = "pg_catalog"
	poolConfig.ConnConfig.RuntimeParams["TimeZone"] = "UTC"
	poolConfig.ConnConfig.RuntimeParams["application_name"] = contract.applicationName
	poolConfig.ConnConfig.RuntimeParams["row_security"] = "on"
	poolConfig.MaxConns = settings.MaxConnections()
	poolConfig.MinConns = settings.MinConnections()
	poolConfig.MaxConnLifetime = settings.MaxConnectionLifetime()
	poolConfig.MaxConnIdleTime = settings.MaxConnectionIdleTime()
	poolConfig.HealthCheckPeriod = settings.HealthCheckPeriod()
	poolConfig.PingTimeout = settings.ConnectTimeout()
}
