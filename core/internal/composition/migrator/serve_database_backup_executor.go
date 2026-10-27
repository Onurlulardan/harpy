package migrator

import (
	"log/slog"
	"net"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func bindDatabaseBackupRuntime(
	process *lifecycle.Process,
	settings config.DatabaseBackupExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	executor schemaLoop,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtime, err := buildDatabaseBackupRuntime(
		settings, info, pool, monitor, logger, telemetryRuntime,
	)
	if err != nil {
		return failDatabaseBackupStartup(
			settings, pool, telemetryRuntime, logger, processcode.InvalidStartup,
		)
	}
	management, err := httpserver.Bind(
		process.StartupContext(), runtime.management, net.ListenConfig{},
	)
	if err != nil {
		return failDatabaseBackupStartup(
			settings, pool, telemetryRuntime, logger, processcode.ServeFailure,
		)
	}
	return serveDatabaseBackupExecutor(
		process, settings, pool, monitor, executor, runtime, management, logger,
	)
}

func serveDatabaseBackupExecutor(
	process *lifecycle.Process,
	settings config.DatabaseBackupExecutor,
	pool *postgres.Pool,
	monitor *health.Monitor,
	executor schemaLoop,
	runtime databaseBackupRuntime,
	management *httpserver.Bound,
	logger *slog.Logger,
) processcode.Code {
	workload, err := startSchemaWorkload(executor, runtime.gate)
	if err != nil {
		_ = management.Close()
		return failDatabaseBackupStartup(
			settings, pool, runtime.telemetryRuntime, logger,
			processcode.InvalidStartup,
		)
	}
	longRunning := settings.LongRunning()
	host, err := processhost.New(processhost.Dependencies{
		Monitor: monitor, Poller: runtime.poller, Management: management,
		CloseAdmission:     runtime.gate.BeginDrain,
		Workload:           workload.lifecycle(),
		CloseDatabase:      pool.Close,
		ForceCloseDatabase: pool.ForceClose,
		ShutdownTelemetry:  runtime.telemetryRuntime.Shutdown,
		DrainTimeout:       longRunning.DrainTimeout(),
		ForceFlushTimeout: longRunning.Runtime().
			Telemetry().ExportTimeout(),
	})
	if err != nil {
		workload.cancel()
		_ = management.Close()
		return failDatabaseBackupStartup(
			settings, pool, runtime.telemetryRuntime, logger,
			processcode.InvalidStartup,
		)
	}
	outcome := host.RunProcess(process)
	if outcome.ExitCode() == processcode.ServeFailure {
		reportFailure(nil, logger, "runtime", "database-backup-executor")
	}
	if outcome.ExitCode() == processcode.DrainForced {
		reportFailure(nil, logger, "drain", "database-backup-executor")
	}
	return outcome.ExitCode()
}
