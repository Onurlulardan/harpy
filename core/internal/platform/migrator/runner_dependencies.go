package migrator

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type RunnerDependencies struct {
	Database         *sql.DB
	TenantDatabase   func() (*sql.DB, error)
	Registry         *migration.ReleaseRegistry
	Release          attestation.VerifiedRelease
	Authority        attestation.VerifiedBootstrapAuthority
	Artifacts        attestation.VerifiedPostgresArtifactSet
	BootstrapReceipt attestation.VerifiedBootstrapReceipt
	ExtensionReceipt attestation.VerifiedExtensionReceipt
	CleanupReceipt   attestation.VerifiedCredentialCleanupReceipt
	ReleaseBinding   schemacontract.ReleaseBinding
	Documents        []migration.Document
	ReleaseFinalizer ReleaseFinalizer
	CleanupTimeout   time.Duration
	Logger           *slog.Logger
}
