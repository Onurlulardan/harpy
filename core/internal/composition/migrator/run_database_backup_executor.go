package migrator

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func runDatabaseBackupExecutorProcess(
	process *lifecycle.Process,
	stderr io.Writer,
) processcode.Code {
	ctx := process.StartupContext()
	settings, err := config.LoadDatabaseBackupExecutor()
	if err != nil {
		reportFailure(stderr, nil, "configuration", "database-backup-executor")
		return processcode.InvalidStartup
	}
	runtimeSettings := settings.LongRunning().Runtime()
	info, err := buildinfo.ReadMigrator(runtimeSettings.Environment())
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "database-backup-executor")
		return processcode.InvalidStartup
	}
	embedded, err := loadSchemaContract()
	if err != nil {
		reportFailure(stderr, nil, "build_contract", "database-backup-executor")
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(stderr, runtimeSettings.LogLevel(), info)
	if err != nil {
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(
		ctx, runtimeSettings.Telemetry(), info, logger,
	)
	if err != nil {
		reportFailure(nil, logger, "configuration", "database-backup-executor")
		return processcode.InvalidStartup
	}
	return startDatabaseBackupExecutor(
		process, settings, info, embedded, logger, telemetryRuntime,
	)
}

func databaseBackupStartupContext(
	process *lifecycle.Process,
	settings config.DatabaseBackupExecutor,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		process.StartupContext(),
		settings.LongRunning().ReadinessTimeout(),
	)
}
