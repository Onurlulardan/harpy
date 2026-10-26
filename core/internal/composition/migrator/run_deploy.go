package migrator

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/deployevidence"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func runDeploy(
	ctx context.Context,
	action deployAction,
	stdout io.Writer,
	stderr io.Writer,
) processcode.Code {
	settings, err := config.LoadMigratorDeploy()
	if err != nil {
		reportFailure(stderr, nil, "configuration", string(action))
		return processcode.InvalidStartup
	}
	info, err := buildinfo.ReadMigrator(settings.Runtime().Environment())
	if err != nil {
		reportFailure(stderr, nil, "build_contract", string(action))
		return processcode.InvalidStartup
	}
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		reportFailure(stderr, nil, "build_contract", string(action))
		return processcode.InvalidStartup
	}
	embedded, err := schemacontract.LoadEmbedded(registry)
	if err != nil {
		reportFailure(stderr, nil, "build_contract", string(action))
		return processcode.InvalidStartup
	}
	logger, err := telemetry.NewLogger(stderr, settings.Runtime().LogLevel(), info)
	if err != nil {
		reportFailure(stderr, nil, "configuration", string(action))
		return processcode.InvalidStartup
	}
	telemetryRuntime, err := telemetry.New(
		ctx, settings.Runtime().Telemetry(), info, logger,
	)
	if err != nil {
		reportFailure(stderr, logger, "configuration", string(action))
		return processcode.InvalidStartup
	}
	return runVerifiedDeploy(
		ctx, action, stdout, stderr, settings, registry, embedded,
		logger, telemetryRuntime,
	)
}

func runVerifiedDeploy(
	ctx context.Context,
	action deployAction,
	stdout io.Writer,
	stderr io.Writer,
	settings config.MigratorDeploy,
	registry *migration.ReleaseRegistry,
	embedded schemacontract.EmbeddedContract,
	logger *slog.Logger,
	telemetryRuntime *telemetry.Runtime,
) processcode.Code {
	database, err := platformmigrator.NewDeployDatabase(settings.Runtime().Postgres())
	if err != nil {
		reportFailure(stderr, logger, "configuration", string(action))
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	evidence, err := deployevidence.Load(
		settings, database.Name(), registry, embedded, time.Now().UTC(),
	)
	if err != nil {
		reportDependencyStage(logger, "release_evidence")
		reportFailure(stderr, logger, "dependency_preflight", string(action))
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	documents, err := deployMigrationDocuments()
	if err != nil {
		reportFailure(stderr, logger, "build_contract", string(action))
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	documents, err = registry.SelectReleaseDocuments(
		evidence.Artifacts().Artifacts().Migration, documents,
	)
	if err != nil {
		reportFailure(stderr, logger, "build_contract", string(action))
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.InvalidStartup
	}
	releaseFinalizer, err := catalogmodule.NewReleaseFinalizerFromPackageSet(
		evidence.Release(),
		settings.CatalogPackageKeyring(),
		evidence.CatalogPackages(),
	)
	if err != nil {
		reportDependencyStage(logger, "catalog_release")
		reportFailure(stderr, logger, "dependency_preflight", string(action))
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	runner, err := platformmigrator.NewRunner(platformmigrator.RunnerDependencies{
		Database: database.SQL(), TenantDatabase: database.OpenTenantDatabase,
		Registry: evidence.Registry(),
		Release:  evidence.Release(), Authority: evidence.Authority(),
		Artifacts: evidence.Artifacts(), BootstrapReceipt: evidence.BootstrapReceipt(),
		ExtensionReceipt: evidence.ExtensionReceipt(), CleanupReceipt: evidence.CleanupReceipt(),
		ReleaseBinding: evidence.ReleaseBinding(),
		Documents:      documents, CleanupTimeout: settings.Runtime().Postgres().CleanupTimeout(),
		Logger: logger, ReleaseFinalizer: releaseFinalizer,
	})
	if err != nil {
		reportDependencyStage(logger, "runner_build")
		reportFailure(stderr, logger, "dependency_preflight", string(action))
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return processcode.DependencyPreflight
	}
	result, err := executeDeploy(ctx, runner, action, randomid.NewUUIDv4)
	if err != nil {
		code, class := classifyDeployFailure(err)
		if class == "dependency_preflight" {
			reportDependencyStage(logger, "database_session")
		}
		if class == "migration" {
			reportFailureCause(stderr, logger, class, string(action), err)
		} else {
			reportFailure(stderr, logger, class, string(action))
		}
		_ = database.Close()
		shutdownDeployTelemetry(settings, telemetryRuntime)
		return code
	}
	if database.Close() != nil || shutdownDeployTelemetry(settings, telemetryRuntime) != nil {
		reportFailure(stderr, logger, "runtime", string(action))
		return processcode.ServeFailure
	}
	if err := result.Write(stdout); err != nil {
		reportFailure(stderr, logger, "runtime", string(action))
		return processcode.ServeFailure
	}
	return processcode.Clean
}

func classifyDeployFailure(err error) (processcode.Code, string) {
	if errors.Is(err, errDeployRuntime) {
		return processcode.ServeFailure, "runtime"
	}
	if errors.Is(err, platformmigrator.ErrDependencyPreflight) {
		return processcode.DependencyPreflight, "dependency_preflight"
	}
	return processcode.DeployMigration, "migration"
}

func shutdownDeployTelemetry(
	settings config.MigratorDeploy,
	runtime *telemetry.Runtime,
) error {
	if runtime == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(
		context.Background(), settings.Runtime().Telemetry().ExportTimeout(),
	)
	defer cancel()
	return runtime.Shutdown(ctx)
}
