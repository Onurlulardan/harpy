package migrator

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/management"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type schemaRuntime struct {
	gate             *admission.Gate
	poller           *health.Poller
	management       *http.Server
	telemetryRuntime *telemetry.Runtime
}

func buildSchemaRuntime(
	settings config.SchemaExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) (schemaRuntime, error) {
	if pool == nil || monitor == nil || logger == nil || telemetryRuntime == nil {
		return schemaRuntime{}, fmt.Errorf("schema executor runtime dependencies are incomplete")
	}
	gate := &admission.Gate{}
	managementHandler, err := management.NewHandler(monitor, info, schemaRuntimeReporter(logger))
	if err != nil {
		return schemaRuntime{}, err
	}
	errorLog, err := telemetry.StandardErrorLog(logger)
	if err != nil {
		return schemaRuntime{}, err
	}
	managementServer, err := httpserver.NewManagement(
		settings.LongRunning().ManagementListenAddress(), managementHandler, errorLog,
	)
	if err != nil {
		return schemaRuntime{}, err
	}
	runtimeSettings := settings.LongRunning().Runtime()
	poller, err := health.NewPoller(
		monitor, runtimeSettings.Postgres().HealthCheckPeriod(),
		runtimeSettings.ReadinessTimeout(), probeSchemaReadiness(pool),
		observeSchemaReadiness(gate),
	)
	if err != nil {
		return schemaRuntime{}, err
	}
	return schemaRuntime{
		gate: gate, poller: poller, management: managementServer,
		telemetryRuntime: telemetryRuntime,
	}, nil
}

func schemaRuntimeReporter(logger *slog.Logger) func(error) {
	return func(error) {
		logger.Error(
			"core migrator runtime failure",
			slog.String("error_class", "runtime"),
			slog.String("command", "schema-executor"),
		)
	}
}
