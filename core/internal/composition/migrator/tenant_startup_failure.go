package migrator

import (
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func failTenantConfiguration(
	settings config.TenantMigrationExecutor,
	runtime *telemetry.Runtime, logger *slog.Logger,
) processcode.Code {
	reportFailure(nil, logger, "configuration", tenantMigrationCommand)
	shutdownSchemaStartup(settings.LongRunning().Runtime(), runtime)
	return processcode.InvalidStartup
}

func failTenantPreflight(
	settings config.TenantMigrationExecutor,
	runtime *telemetry.Runtime, logger *slog.Logger,
	startupStage string,
) processcode.Code {
	reportFailureStage(
		nil, logger, "dependency_preflight", tenantMigrationCommand, startupStage,
	)
	shutdownSchemaStartup(settings.LongRunning().Runtime(), runtime)
	return processcode.DependencyPreflight
}

func failTenantStartup(
	settings config.TenantMigrationExecutor,
	pool *postgres.Pool,
	database *platformmigrator.TenantExecutorDatabase,
	runtime *telemetry.Runtime,
	logger *slog.Logger,
	startupStage string,
) processcode.Code {
	reportFailureStage(
		nil, logger, "build_contract", tenantMigrationCommand, startupStage,
	)
	if pool != nil {
		pool.Close()
	}
	_ = database.Close()
	shutdownSchemaStartup(settings.LongRunning().Runtime(), runtime)
	return processcode.InvalidStartup
}
