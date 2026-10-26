package migrator

import (
	"database/sql"
	"fmt"
	"log/slog"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

type TenantExecutorDependencies struct {
	Database   *sql.DB
	OpenTenant func() (*sql.DB, error)
	Registry   *migration.ReleaseRegistry
	Release    attestation.VerifiedRelease
	Documents  []migration.Document
	Logger     *slog.Logger
	NewID      func() (uuid.UUID, error)
}

type TenantExecutor struct {
	database *sql.DB
	target   *TenantTargetExecutor
	release  attestation.ReleaseManifest
	newID    func() (uuid.UUID, error)
}

func NewTenantExecutor(
	dependencies TenantExecutorDependencies,
) (*TenantExecutor, error) {
	if dependencies.Database == nil || dependencies.OpenTenant == nil ||
		dependencies.Registry == nil || dependencies.Logger == nil ||
		dependencies.NewID == nil {
		return nil, fmt.Errorf("tenant migration runner dependencies are incomplete")
	}
	release := dependencies.Release.Value()
	if release.ReleaseID == "" || release.MigrationManifestSHA256 == "" {
		return nil, fmt.Errorf("verified tenant migration release is required")
	}
	artifact, err := dependencies.Registry.MigrationArtifact(
		release.MigrationManifestSHA256,
	)
	if err != nil {
		return nil, err
	}
	documents, err := dependencies.Registry.SelectReleaseDocuments(
		artifact, dependencies.Documents,
	)
	if err != nil {
		return nil, err
	}
	catalog, err := migration.NewTenantCatalog(dependencies.Registry, artifact,
		parseDatabaseOrigin(release.DatabaseOrigin), documents)
	if err != nil {
		return nil, err
	}
	files, err := catalog.FS(tenantMigrationScope)
	if err != nil {
		return nil, err
	}
	selected, err := catalog.Selected(tenantMigrationScope)
	if err != nil {
		return nil, err
	}
	target, err := NewTenantTargetExecutor(TenantTargetDependencies{
		Database: dependencies.OpenTenant, Files: files,
		Selected:             selected,
		HistorySupersessions: catalog.HistorySupersessions(tenantMigrationScope),
		Logger:               dependencies.Logger,
	})
	if err != nil {
		return nil, err
	}
	return &TenantExecutor{database: dependencies.Database, target: target,
		release: release, newID: dependencies.NewID}, nil
}

func parseDatabaseOrigin(value string) postgresv1.DatabaseOrigin {
	return postgresv1.DatabaseOrigin(value)
}
