package migrator

import (
	"context"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	schemaexecutorstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func startSchemaExecutor(
	process *lifecycle.Process,
	settings config.SchemaExecutor,
	info buildinfo.Info,
	embedded schemacontract.EmbeddedContract,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	return startSchemaExecutorWithOpen(
		process, settings, info, embedded, logger, telemetryRuntime, postgres.Open,
	)
}

type schemaPoolOpener func(
	context.Context,
	config.Postgres,
	postgres.Profile,
	schemacontract.EmbeddedContract,
) (*postgres.Pool, error)

func startSchemaExecutorWithOpen(
	process *lifecycle.Process,
	settings config.SchemaExecutor,
	info buildinfo.Info,
	embedded schemacontract.EmbeddedContract,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
	openPool schemaPoolOpener,
) processcode.Code {
	runtimeSettings := settings.LongRunning().Runtime()
	startupContext, cancelStartup := context.WithTimeout(
		process.StartupContext(), runtimeSettings.ReadinessTimeout(),
	)
	pool, err := openPool(
		startupContext, runtimeSettings.Postgres(), postgres.ProfileSchemaExecutor, embedded,
	)
	if err != nil {
		cancelStartup()
		reportFailure(nil, logger, "dependency_preflight", "schema-executor")
		shutdownSchemaStartup(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	initial := probeSchemaReadiness(pool)(startupContext)
	cancelStartup()
	if !schemaReady(initial) {
		reportFailure(nil, logger, "dependency_preflight", "schema-executor")
		pool.Close()
		shutdownSchemaStartup(runtimeSettings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	persistence, err := schemaexecutorstore.Bind(pool)
	if err != nil {
		return failSchemaStartup(settings, pool, telemetryRuntime, logger)
	}
	executorID, err := randomid.NewUUIDv4()
	if err != nil {
		return failSchemaStartup(settings, pool, telemetryRuntime, logger)
	}
	executor, err := schemaexecutor.New(schemaexecutor.Dependencies{
		Persistence: persistence, ExecutorID: executorID,
		NewLeaseID: randomid.NewUUIDv4,
	})
	if err != nil {
		return failSchemaStartup(settings, pool, telemetryRuntime, logger)
	}
	monitor, err := health.NewMonitor(
		info.Service(), runtimeSettings.Postgres().HealthCheckPeriod(), nil,
	)
	if err != nil {
		return failSchemaStartup(settings, pool, telemetryRuntime, logger)
	}
	return bindSchemaExecutor(
		process, settings, info, pool, monitor, executor, logger, telemetryRuntime,
	)
}

func failSchemaStartup(
	settings config.SchemaExecutor,
	pool *postgres.Pool,
	telemetryRuntime *telemetry.Runtime,
	logger *slog.Logger,
) processcode.Code {
	reportFailure(nil, logger, "build_contract", "schema-executor")
	pool.Close()
	shutdownSchemaStartup(settings.LongRunning().Runtime(), telemetryRuntime)
	return processcode.InvalidStartup
}
