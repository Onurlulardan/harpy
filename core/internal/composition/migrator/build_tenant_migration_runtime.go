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

type tenantMigrationRuntime struct {
	gate             *admission.Gate
	poller           *health.Poller
	management       *http.Server
	telemetryRuntime *telemetry.Runtime
}

func buildTenantMigrationRuntime(
	settings config.TenantMigrationExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) (tenantMigrationRuntime, error) {
	if pool == nil || monitor == nil || logger == nil || telemetryRuntime == nil {
		return tenantMigrationRuntime{}, fmt.Errorf(
			"tenant migration runtime dependencies are incomplete",
		)
	}
	gate := &admission.Gate{}
	handler, err := management.NewHandler(
		monitor, info, tenantMigrationRuntimeReporter(logger),
	)
	if err != nil {
		return tenantMigrationRuntime{}, err
	}
	errorLog, err := telemetry.StandardErrorLog(logger)
	if err != nil {
		return tenantMigrationRuntime{}, err
	}
	server, err := httpserver.NewManagement(
		settings.LongRunning().ManagementListenAddress(), handler, errorLog,
	)
	if err != nil {
		return tenantMigrationRuntime{}, err
	}
	runtimeSettings := settings.LongRunning().Runtime()
	poller, err := health.NewPoller(monitor,
		runtimeSettings.Postgres().HealthCheckPeriod(),
		runtimeSettings.ReadinessTimeout(), probeSchemaReadiness(pool),
		observeSchemaReadiness(gate))
	if err != nil {
		return tenantMigrationRuntime{}, err
	}
	return tenantMigrationRuntime{gate: gate, poller: poller,
		management: server, telemetryRuntime: telemetryRuntime}, nil
}

func tenantMigrationRuntimeReporter(logger *slog.Logger) func(error) {
	return func(error) {
		logger.Error("core migrator runtime failure",
			slog.String("error_class", "runtime"),
			slog.String("command", tenantMigrationCommand))
	}
}
