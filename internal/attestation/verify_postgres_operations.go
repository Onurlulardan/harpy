package attestation

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func VerifyBootstrapOperationArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.BootstrapOperation], error) {
	return verifyPostgresArtifact[postgresv1.BootstrapOperation](
		document, postgresv1.MaximumBootstrapOperationBytes,
	)
}

func VerifyExtensionOperationArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.ExtensionOperation], error) {
	return verifyPostgresArtifact[postgresv1.ExtensionOperation](
		document, postgresv1.MaximumExtensionOperationBytes,
	)
}

func VerifyCredentialCleanupOperationArtifact(
	document []byte,
) (VerifiedPostgresArtifact[postgresv1.CredentialCleanupOperation], error) {
	return verifyPostgresArtifact[postgresv1.CredentialCleanupOperation](
		document, postgresv1.MaximumCredentialCleanupOperationBytes,
	)
}
