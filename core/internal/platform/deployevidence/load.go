package deployevidence

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

// Load verifies every deploy artifact before a network connection is opened.
func Load(
	settings config.MigratorDeploy,
	targetDatabase string,
	registry *migration.ReleaseRegistry,
	embedded schemacontract.EmbeddedContract,
	now time.Time,
) (Evidence, error) {
	verifiedAuthority, err := verifyReleaseAuthority(
		settings.Verification(), targetDatabase, now,
	)
	if err != nil {
		return Evidence{}, err
	}
	packageSet, err := verifyCatalogPackageSet(
		settings, verifiedAuthority,
	)
	if err != nil {
		return Evidence{}, err
	}
	artifacts, err := verifyPostgresArtifacts(settings.Artifacts())
	if err != nil {
		return Evidence{}, err
	}
	boundArtifacts, err := attestation.BindPostgresArtifactSet(
		verifiedAuthority.release, verifiedAuthority.authority, artifacts,
	)
	if err != nil {
		return Evidence{}, fmt.Errorf("PostgreSQL artifact binding failed")
	}
	receipts, err := verifyRuntimeReceipts(
		settings.Verification(), verifiedAuthority.root, now,
	)
	if err != nil {
		return Evidence{}, err
	}
	binding, err := schemacontract.BindRelease(
		embedded, verifiedAuthority.release,
		boundArtifacts.Artifacts().Migration,
		boundArtifacts.Artifacts().Extension,
	)
	if err != nil {
		return Evidence{}, fmt.Errorf("binary schema contract binding failed")
	}
	return Evidence{
		registry: registry, release: verifiedAuthority.release,
		authority: verifiedAuthority.authority, artifacts: boundArtifacts,
		bootstrapReceipt: receipts.bootstrap, extensionReceipt: receipts.extension,
		cleanupReceipt: receipts.cleanup, catalogPackages: packageSet,
		releaseBinding: binding,
	}, nil
}
