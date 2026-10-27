package migrator

import (
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/processhost"
)

func serveSchemaExecutor(
	process *lifecycle.Process,
	settings config.SchemaExecutor,
	pool *postgres.Pool,
	monitor *health.Monitor,
	executor schemaLoop,
	runtime schemaRuntime,
	management *httpserver.Bound,
	logger *slog.Logger,
) processcode.Code {
	workload, err := startSchemaWorkload(executor, runtime.gate)
	if err != nil {
		_ = management.Close()
		return failSchemaStartup(settings, pool, runtime.telemetryRuntime, logger)
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
		ForceFlushTimeout:  longRunning.Runtime().Telemetry().ExportTimeout(),
	})
	if err != nil {
		workload.cancel()
		_ = management.Close()
		return failSchemaStartup(settings, pool, runtime.telemetryRuntime, logger)
	}
	outcome := host.RunProcess(process)
	if outcome.ExitCode() == processcode.ServeFailure {
		reportFailure(nil, logger, "runtime", "schema-executor")
	}
	if outcome.ExitCode() == processcode.DrainForced {
		reportFailure(nil, logger, "drain", "schema-executor")
	}
	return outcome.ExitCode()
}
