package migrator

import (
	"database/sql"
	"fmt"
	"log/slog"
	"testing/fstest"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

type TenantExecutorDependencies struct {
	Database   *sql.DB
	OpenTenant func() (*sql.DB, error)
	Release    attestation.VerifiedRelease
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
		dependencies.Logger == nil || dependencies.NewID == nil {
		return nil, fmt.Errorf("tenant migration runner dependencies are incomplete")
	}
	release := dependencies.Release.Value()
	if release.ReleaseID == "" || release.MigrationManifestSHA256 == "" {
		return nil, fmt.Errorf("verified tenant migration release is required")
	}
	target, err := NewTenantTargetExecutor(TenantTargetDependencies{
		Database: dependencies.OpenTenant,
		Files:    fstest.MapFS{},
		Selected: []postgresv1.SelectedMigration{tenantBaselineSelection()},
		Logger:   dependencies.Logger,
	})
	if err != nil {
		return nil, err
	}
	return &TenantExecutor{database: dependencies.Database, target: target,
		release: release, newID: dependencies.NewID}, nil
}

func tenantBaselineSelection() postgresv1.SelectedMigration {
	return postgresv1.SelectedMigration{
		Owner: "platform_migrator", Scope: tenantMigrationScope,
		Version: 20261028000001, TransactionMode: "transactional",
		Origin: postgresv1.DatabaseOriginFresh,
		Path:   "20261028000001_tenant_baseline.fresh.sql",
	}
}
