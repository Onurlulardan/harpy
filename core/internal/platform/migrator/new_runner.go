package migrator

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func NewRunner(dependencies RunnerDependencies) (*Runner, error) {
	if dependencies.Database == nil || dependencies.Registry == nil ||
		dependencies.Logger == nil {
		reportRunnerBuildStage(dependencies.Logger, "dependencies")
		return nil, fmt.Errorf("migrator dependencies are incomplete")
	}
	if dependencies.CleanupTimeout != requiredCleanupTimeout {
		reportRunnerBuildStage(dependencies.Logger, "cleanup_timeout")
		return nil, fmt.Errorf("migrator cleanup timeout must be exactly %s", requiredCleanupTimeout)
	}
	if !dependencies.ReleaseBinding.IsVerified() {
		reportRunnerBuildStage(dependencies.Logger, "embedded_release_binding")
		return nil, fmt.Errorf("embedded schema contract release binding is required")
	}
	artifacts, err := attestation.BindPostgresArtifactSet(
		dependencies.Release,
		dependencies.Authority,
		dependencies.Artifacts.Artifacts(),
	)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "postgres_artifact_binding")
		return nil, fmt.Errorf("bind migrator release evidence: %w", err)
	}
	catalog, err := migration.NewCatalog(
		dependencies.Registry,
		artifacts,
		dependencies.Documents,
	)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "migration_catalog")
		return nil, fmt.Errorf("build verified migration catalog: %w", err)
	}
	files, err := catalog.FS(platformMigrationScope)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "platform_filesystem")
		return nil, err
	}
	selected, err := catalog.Selected(platformMigrationScope)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "platform_selection")
		return nil, err
	}
	if err := validatePlatformSelection(selected); err != nil {
		reportRunnerBuildStage(dependencies.Logger, "platform_selection_validation")
		return nil, err
	}
	historyAliases, err := indexHistorySupersessions(
		selected, catalog.HistorySupersessions(platformMigrationScope),
	)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "history_supersessions")
		return nil, err
	}
	tenantFiles, err := catalog.FS(tenantMigrationScope)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "tenant_filesystem")
		return nil, err
	}
	tenantSelected, err := catalog.Selected(tenantMigrationScope)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "tenant_selection")
		return nil, err
	}
	if err := validateTenantSelection(tenantSelected); err != nil {
		reportRunnerBuildStage(dependencies.Logger, "tenant_selection_validation")
		return nil, err
	}
	tenantAliases, err := indexHistorySupersessions(
		tenantSelected, catalog.HistorySupersessions(tenantMigrationScope),
	)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "tenant_history_supersessions")
		return nil, err
	}
	if dependencies.TenantDatabase == nil {
		reportRunnerBuildStage(dependencies.Logger, "tenant_database")
		return nil, fmt.Errorf("tenant migration dependencies are required")
	}
	boundary, err := selectReleaseGenerationBoundary(selected, tenantSelected)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "release_generation_boundary")
		return nil, err
	}
	binding, err := schemacontract.BindRelease(
		dependencies.ReleaseBinding.Contract(), dependencies.Release,
		artifacts.Artifacts().Migration, artifacts.Artifacts().Extension,
	)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "schema_contract_binding")
		return nil, fmt.Errorf("bind embedded schema contract to migrator release: %w", err)
	}
	if !binding.IsVerified() {
		reportRunnerBuildStage(dependencies.Logger, "verified_schema_contract_binding")
		return nil, fmt.Errorf("embedded schema contract release binding is not verified")
	}
	binaryContract := binding.Contract().Value()
	if binaryContract.MinimumDatabaseGeneration > boundary.databaseGeneration ||
		binaryContract.RuntimeGeneration < boundary.minimumRuntimeGeneration {
		reportRunnerBuildStage(dependencies.Logger, "binary_generation_admission")
		return nil, fmt.Errorf("binary schema generation cannot admit selected migrations")
	}
	identity, evidence, err := bindFoundationEvidence(dependencies, artifacts, boundary)
	if err != nil {
		reportRunnerBuildStage(dependencies.Logger, "foundation_evidence")
		return nil, err
	}
	return &Runner{
		database: dependencies.Database, platformFiles: files,
		selected: selected, historyAliases: historyAliases,
		tenantDatabase: dependencies.TenantDatabase,
		tenantFiles:    tenantFiles, tenantSelected: tenantSelected,
		tenantAliases: tenantAliases,
		identity:      identity, evidence: evidence,
		schemaAuthority:  releaseSchemaAuthority{binding: binding},
		releaseFinalizer: admitReleaseFinalizer(selected, dependencies.ReleaseFinalizer),
		cleanupTimeout:   dependencies.CleanupTimeout, logger: dependencies.Logger,
	}, nil
}

func reportRunnerBuildStage(logger *slog.Logger, stage string) {
	if logger == nil || stage == "" {
		return
	}
	logger.Debug(
		"core migrator runner build stage failed",
		slog.String("runner_build_stage", stage),
	)
}
