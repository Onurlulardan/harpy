package migrator

import (
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func serveTenantMigrationExecutor(
	process *lifecycle.Process,
	settings config.TenantMigrationExecutor,
	pool *postgres.Pool,
	database *platformmigrator.TenantExecutorDatabase,
	monitor *health.Monitor,
	executor *tenantexec.Executor,
	runtime tenantMigrationRuntime,
	management *httpserver.Bound,
	logger *slog.Logger,
) processcode.Code {
	workload, err := startTenantMigrationWorkload(executor, runtime.gate)
	if err != nil {
		_ = management.Close()
		return failTenantStartup(settings, pool, database,
			runtime.telemetryRuntime, logger, "workload_start")
	}
	longRunning := settings.LongRunning()
	host, err := processhost.New(processhost.Dependencies{
		Monitor: monitor, Poller: runtime.poller, Management: management,
		CloseAdmission:     workload.gate.BeginDrain,
		Workload:           workload.lifecycle(),
		CloseDatabase:      func() { pool.Close(); _ = database.Close() },
		ForceCloseDatabase: func() { pool.ForceClose(); _ = database.Close() },
		ShutdownTelemetry:  runtime.telemetryRuntime.Shutdown,
		DrainTimeout:       longRunning.DrainTimeout(),
		ForceFlushTimeout:  longRunning.Runtime().Telemetry().ExportTimeout(),
	})
	if err != nil {
		workload.cancel()
		_ = management.Close()
		return failTenantStartup(settings, pool, database,
			runtime.telemetryRuntime, logger, "process_host")
	}
	outcome := host.RunProcess(process)
	if outcome.ExitCode() == processcode.ServeFailure {
		reportFailure(nil, logger, "runtime", tenantMigrationCommand)
	}
	if outcome.ExitCode() == processcode.DrainForced {
		reportFailure(nil, logger, "drain", tenantMigrationCommand)
	}
	return outcome.ExitCode()
}
