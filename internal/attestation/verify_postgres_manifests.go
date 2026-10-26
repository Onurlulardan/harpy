package attestation

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func VerifyMigrationManifestArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.MigrationManifest], error) {
	return verifyPostgresArtifact[postgresv1.MigrationManifest](
		document, postgresv1.MaximumMigrationManifestBytes,
	)
}

func VerifyExtensionManifestArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.ExtensionManifest], error) {
	return verifyPostgresArtifact[postgresv1.ExtensionManifest](
		document, postgresv1.MaximumExtensionManifestBytes,
	)
}

func VerifyRolesManifestArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.RolesManifest], error) {
	return verifyPostgresArtifact[postgresv1.RolesManifest](
		document, postgresv1.MaximumRolesManifestBytes,
	)
}

func VerifyPrivilegesManifestArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.PrivilegesManifest], error) {
	return verifyPostgresArtifact[postgresv1.PrivilegesManifest](
		document, postgresv1.MaximumPrivilegesManifestBytes,
	)
}

func VerifyOwnershipManifestArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], error) {
	return verifyPostgresArtifact[postgresv1.OwnershipManifest](
		document, postgresv1.MaximumOwnershipManifestBytes,
	)
}
