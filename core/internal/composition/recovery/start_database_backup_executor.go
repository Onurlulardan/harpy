package recovery

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func startDatabaseBackupExecutor(
	process *lifecycle.Process,
	settings config.DatabaseBackupExecutor,
	info buildinfo.Info,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtimeSettings := settings.LongRunning().Runtime()
	startup, cancel := databaseBackupStartupContext(process, settings)
	defer cancel()
	pool, err := postgres.Open(
		startup, runtimeSettings.Postgres(),
		postgres.ProfileDatabaseBackupExecutor,
	)
	if err != nil {
		return failDatabaseBackupStartup(
			settings, nil, telemetryRuntime, logger, processcode.DependencyPreflight, err,
		)
	}
	if !schemaReady(probeSchemaReadiness(pool)(startup)) {
		return failDatabaseBackupStartup(
			settings, pool, telemetryRuntime, logger, processcode.DependencyPreflight,
			fmt.Errorf("database backup schema contract is not ready"),
		)
	}
	executor, err := platformops.NewDatabaseBackupExecutor(
		startup, pool, settings,
		platformops.RecoveryFactories{
			Documents:         documents.NewRecovery,
			Targets:           tenancy.NewRecoveryTargets,
			IAMEvidence:       iam.NewRecoveryEvidence,
			AuthorityEvidence: platformauthority.NewRecoveryEvidence,
		},
	)
	if err != nil {
		return failDatabaseBackupStartup(
			settings, pool, telemetryRuntime, logger, processcode.InvalidStartup, err,
		)
	}
	defer executor.Close()
	monitor, err := health.NewMonitor(
		info.Service(), runtimeSettings.Postgres().HealthCheckPeriod(), nil,
	)
	if err != nil {
		return failDatabaseBackupStartup(
			settings, pool, telemetryRuntime, logger, processcode.InvalidStartup, err,
		)
	}
	return bindDatabaseBackupRuntime(
		process, settings, info, pool, monitor, executor, logger, telemetryRuntime,
	)
}

func failDatabaseBackupStartup(
	settings config.DatabaseBackupExecutor,
	pool *postgres.Pool,
	telemetryRuntime *telemetry.Runtime,
	logger *slog.Logger,
	code processcode.Code,
	cause error,
) processcode.Code {
	failureClass := "build_contract"
	if code == processcode.DependencyPreflight {
		failureClass = "dependency_preflight"
	}
	reportFailureCause(nil, logger, failureClass, "database-backup-executor", cause)
	if pool != nil {
		pool.Close()
	}
	if telemetryRuntime != nil {
		runtimeSettings := settings.LongRunning().Runtime()
		ctx, cancel := context.WithTimeout(
			context.Background(), runtimeSettings.Telemetry().ExportTimeout(),
		)
		defer cancel()
		_ = telemetryRuntime.Shutdown(ctx)
	}
	return code
}
