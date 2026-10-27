package migrator

import (
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func bindFoundationEvidence(
	dependencies RunnerDependencies,
	artifacts attestation.VerifiedPostgresArtifactSet,
	boundary releaseGenerationBoundary,
) (runIdentity, foundationEvidence, error) {
	release := dependencies.Release.Value()
	authority := dependencies.Authority.Value()
	bootstrap := dependencies.BootstrapReceipt.Value()
	extension := dependencies.ExtensionReceipt.Value()
	cleanup := dependencies.CleanupReceipt.Value()
	if err := validateReceiptBindings(
		release, dependencies.Release.ManifestSHA256(), authority,
		artifacts, bootstrap, extension, cleanup,
	); err != nil {
		return runIdentity{}, foundationEvidence{}, err
	}
	identity := runIdentity{
		releaseID: release.ReleaseID, environmentID: authority.EnvironmentID,
		databaseName: artifacts.DatabaseName(), databaseOrigin: artifacts.DatabaseOrigin(),
		databaseOwner: release.DatabaseOwner, databaseEncoding: release.DatabaseEncoding,
		databaseLocaleProvider: release.DatabaseLocaleProvider,
		databaseCollation:      release.DatabaseCollation, databaseCType: release.DatabaseCType,
		dataChecksums:            release.DataChecksums,
		manifestSHA256:           artifacts.Artifacts().Migration.DigestSHA256(),
		extensionSHA256:          artifacts.Artifacts().Extension.DigestSHA256(),
		rolesSHA256:              artifacts.Artifacts().Roles.DigestSHA256(),
		privilegesSHA256:         artifacts.Artifacts().Privileges.DigestSHA256(),
		ownershipSHA256:          artifacts.Artifacts().Ownership.Root().DigestSHA256(),
		authoritySHA256:          dependencies.Authority.ManifestSHA256(),
		databaseSystemID:         bootstrap.Payload.DatabaseSystemIdentifier,
		databaseOID:              bootstrap.Payload.DatabaseOID,
		databaseGeneration:       boundary.databaseGeneration,
		minimumRuntimeGeneration: boundary.minimumRuntimeGeneration,
	}
	evidence, err := buildFoundationEvidence(
		artifacts, bootstrap, extension, cleanup,
	)
	if err != nil {
		return runIdentity{}, foundationEvidence{}, err
	}
	return identity, evidence, nil
}
