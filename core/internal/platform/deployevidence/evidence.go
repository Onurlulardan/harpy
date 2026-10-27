package deployevidence

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

// Evidence is the verified, release-bound input set accepted by the migrator.
type Evidence struct {
	registry         *migration.ReleaseRegistry
	release          attestation.VerifiedRelease
	authority        attestation.VerifiedBootstrapAuthority
	artifacts        attestation.VerifiedPostgresArtifactSet
	bootstrapReceipt attestation.VerifiedBootstrapReceipt
	extensionReceipt attestation.VerifiedExtensionReceipt
	cleanupReceipt   attestation.VerifiedCredentialCleanupReceipt
	catalogPackages  attestation.VerifiedCatalogPackageSet
	releaseBinding   schemacontract.ReleaseBinding
}

func (e Evidence) Registry() *migration.ReleaseRegistry { return e.registry }

func (e Evidence) Release() attestation.VerifiedRelease { return e.release }

func (e Evidence) Authority() attestation.VerifiedBootstrapAuthority { return e.authority }

func (e Evidence) Artifacts() attestation.VerifiedPostgresArtifactSet { return e.artifacts }

func (e Evidence) BootstrapReceipt() attestation.VerifiedBootstrapReceipt {
	return e.bootstrapReceipt
}

func (e Evidence) ExtensionReceipt() attestation.VerifiedExtensionReceipt {
	return e.extensionReceipt
}

func (e Evidence) CleanupReceipt() attestation.VerifiedCredentialCleanupReceipt {
	return e.cleanupReceipt
}

func (e Evidence) CatalogPackages() attestation.VerifiedCatalogPackageSet {
	return e.catalogPackages
}

func (e Evidence) ReleaseBinding() schemacontract.ReleaseBinding { return e.releaseBinding }
