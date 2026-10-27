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
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func bindSchemaExecutor(
	process *lifecycle.Process,
	settings config.SchemaExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	executor schemaLoop,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtime, err := buildSchemaRuntime(
		settings, info, pool, monitor, logger, telemetryRuntime,
	)
	if err != nil {
		return failSchemaStartup(settings, pool, telemetryRuntime, logger)
	}
	management, err := httpserver.Bind(
		process.StartupContext(), runtime.management, net.ListenConfig{},
	)
	if err != nil {
		reportFailure(nil, logger, "runtime", "schema-executor")
		pool.Close()
		shutdownSchemaStartup(settings.LongRunning().Runtime(), telemetryRuntime)
		return processcode.ServeFailure
	}
	return serveSchemaExecutor(
		process, settings, pool, monitor, executor, runtime, management, logger,
	)
}
