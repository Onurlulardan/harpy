package migrator

import (
	"log/slog"
	"net"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func bindTenantMigrationExecutor(
	process *lifecycle.Process,
	settings config.TenantMigrationExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	database *platformmigrator.TenantExecutorDatabase,
	monitor *health.Monitor,
	executor *tenantexec.Executor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtime, err := buildTenantMigrationRuntime(
		settings, info, pool, monitor, logger, telemetryRuntime,
	)
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "runtime_build",
		)
	}
	management, err := httpserver.Bind(
		process.StartupContext(), runtime.management, net.ListenConfig{},
	)
	if err != nil {
		reportFailure(nil, logger, "runtime", tenantMigrationCommand)
		pool.Close()
		_ = database.Close()
		shutdownSchemaStartup(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.ServeFailure
	}
	return serveTenantMigrationExecutor(process, settings, pool, database,
		monitor, executor, runtime, management, logger)
}
