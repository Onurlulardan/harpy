package migrator

import (
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/deployevidence"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

const tenantMigrationCommand = "tenant-migration-executor"

func runTenantMigrationExecutorProcess(
	process *lifecycle.Process, stderr io.Writer,
) processcode.Code {
	ctx := process.StartupContext()
	settings, err := config.LoadTenantMigrationExecutor()
	if err != nil {
		reportFailure(stderr, nil, "configuration", tenantMigrationCommand)
		return processcode.InvalidStartup
	}
	runtimeSettings := settings.LongRunning().Runtime()
	info, err := buildinfo.ReadMigrator(runtimeSettings.Environment())
	if err != nil {
		reportFailureStage(
			stderr, nil, "build_contract", tenantMigrationCommand, "build_info",
		)
		return processcode.InvalidStartup
	}
	documents, err := deployMigrationDocuments()
	if err != nil {
		reportFailureStage(
			stderr, nil, "build_contract", tenantMigrationCommand, "schema_contract",
		)
		return processcode.InvalidStartup
	}
	embedded, registry, err := schemacontract.LoadStartup(documents)
	if err != nil {
		reportFailureStage(
			stderr, nil, "build_contract", tenantMigrationCommand, "schema_contract",
		)
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(stderr, runtimeSettings.LogLevel(), info)
	if err != nil {
		reportFailure(stderr, nil, "configuration", tenantMigrationCommand)
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(
		ctx, runtimeSettings.Telemetry(), info, logger,
	)
	if err != nil {
		reportFailure(nil, logger, "configuration", tenantMigrationCommand)
		return processcode.InvalidStartup
	}
	database, err := platformmigrator.NewTenantExecutorDatabase(
		runtimeSettings.Postgres(),
	)
	if err != nil {
		return failTenantConfiguration(settings, telemetryRuntime, logger)
	}
	release, err := deployevidence.LoadTenantMigrationRelease(
		settings.Release(), database.Name(), time.Now().UTC(),
	)
	if err != nil {
		_ = database.Close()
		return failTenantPreflight(
			settings, telemetryRuntime, logger, "release_evidence",
		)
	}
	return startTenantMigrationExecutor(process, settings, info, embedded,
		registry, release, documents, database, logger, telemetryRuntime)
}
