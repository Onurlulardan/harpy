package migrator

import (
	"context"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	tenantmigrationstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/tenantmigrationexecutor/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func startTenantMigrationExecutor(
	process *lifecycle.Process,
	settings config.TenantMigrationExecutor,
	info buildinfo.Info,
	embedded schemacontract.EmbeddedContract,
	registry *migration.ReleaseRegistry,
	release attestation.VerifiedRelease,
	documents []migration.Document,
	database *platformmigrator.TenantExecutorDatabase,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	runtimeSettings := settings.LongRunning().Runtime()
	startup, cancel := context.WithTimeout(
		process.StartupContext(), runtimeSettings.ReadinessTimeout(),
	)
	defer cancel()
	pool, err := postgres.Open(startup, runtimeSettings.Postgres(),
		postgres.ProfileTenantMigrationExecutor, embedded)
	if err != nil || database.SQL().PingContext(startup) != nil {
		if pool != nil {
			pool.Close()
		}
		_ = database.Close()
		return failTenantPreflight(
			settings, telemetryRuntime, logger, "database_connect",
		)
	}
	if !schemaReady(probeSchemaReadiness(pool)(startup)) {
		pool.Close()
		_ = database.Close()
		return failTenantPreflight(
			settings, telemetryRuntime, logger, "schema_readiness",
		)
	}
	store, err := tenantmigrationstore.Bind(pool)
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "store_bind",
		)
	}
	migrator, err := platformmigrator.NewTenantExecutor(
		platformmigrator.TenantExecutorDependencies{Database: database.SQL(),
			OpenTenant: database.OpenTenant, Registry: registry, Release: release,
			Documents: documents, Logger: logger, NewID: tenantexec.NewID},
	)
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "migrator_build",
		)
	}
	manifest := release.Value()
	executorRelease, err := tenantexec.NewRelease(manifest.ReleaseID,
		string(settings.DeploymentMode()), manifest.RuntimeGeneration,
		manifest.MigrationManifestSHA256)
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "release_build",
		)
	}
	executor, err := tenantexec.New(tenantexec.Dependencies{
		Store: store, Migrator: migrator, NewID: tenantexec.NewID,
		Release: executorRelease,
	})
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "executor_build",
		)
	}
	monitor, err := health.NewMonitor(
		info.Service(), runtimeSettings.Postgres().HealthCheckPeriod(), nil,
	)
	if err != nil {
		return failTenantStartup(
			settings, pool, database, telemetryRuntime, logger, "health_monitor",
		)
	}
	return bindTenantMigrationExecutor(process, settings, info, pool, database,
		monitor, executor, logger, telemetryRuntime)
}
