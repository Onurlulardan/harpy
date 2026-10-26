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

type databaseBackupRuntime struct {
	gate             *admission.Gate
	poller           *health.Poller
	management       *http.Server
	telemetryRuntime *telemetry.Runtime
}

func buildDatabaseBackupRuntime(
	settings config.DatabaseBackupExecutor,
	info buildinfo.Info,
	pool *postgres.Pool,
	monitor *health.Monitor,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) (databaseBackupRuntime, error) {
	if pool == nil || monitor == nil || logger == nil || telemetryRuntime == nil {
		return databaseBackupRuntime{}, fmt.Errorf(
			"database backup runtime dependencies are incomplete",
		)
	}
	gate := &admission.Gate{}
	handler, err := management.NewHandler(
		monitor, info, databaseBackupRuntimeReporter(logger),
	)
	if err != nil {
		return databaseBackupRuntime{}, err
	}
	errorLog, err := telemetry.StandardErrorLog(logger)
	if err != nil {
		return databaseBackupRuntime{}, err
	}
	server, err := httpserver.NewManagement(
		settings.LongRunning().ManagementListenAddress(), handler, errorLog,
	)
	if err != nil {
		return databaseBackupRuntime{}, err
	}
	runtimeSettings := settings.LongRunning().Runtime()
	poller, err := health.NewPoller(
		monitor, runtimeSettings.Postgres().HealthCheckPeriod(),
		settings.LongRunning().ReadinessTimeout(),
		probeSchemaReadiness(pool), observeSchemaReadiness(gate),
	)
	if err != nil {
		return databaseBackupRuntime{}, err
	}
	return databaseBackupRuntime{
		gate: gate, poller: poller, management: server,
		telemetryRuntime: telemetryRuntime,
	}, nil
}

func databaseBackupRuntimeReporter(logger *slog.Logger) func(error) {
	return func(error) {
		logger.Error(
			"database backup executor runtime failure",
			slog.String("error_class", "runtime"),
			slog.String("command", "database-backup-executor"),
		)
	}
}
