package deployevidence

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type verifiedPostgresOperations struct {
	bootstrap attestation.VerifiedPostgresArtifact[postgresv1.BootstrapOperation]
	extension attestation.VerifiedPostgresArtifact[postgresv1.ExtensionOperation]
	cleanup   attestation.VerifiedPostgresArtifact[postgresv1.CredentialCleanupOperation]
}

func verifyPostgresOperations(
	settings config.MigrationArtifacts,
) (verifiedPostgresOperations, error) {
	bootstrapDocument, err := readArtifact(
		settings.BootstrapOperationPath().Value(), postgresv1.MaximumBootstrapOperationBytes,
		"bootstrap operation",
	)
	if err != nil {
		return verifiedPostgresOperations{}, err
	}
	bootstrap, err := attestation.VerifyBootstrapOperationArtifact(bootstrapDocument.bytes)
	if err != nil {
		return verifiedPostgresOperations{}, fmt.Errorf("bootstrap operation verification failed")
	}
	extensionDocument, err := readArtifact(
		settings.ExtensionOperationPath().Value(), postgresv1.MaximumExtensionOperationBytes,
		"extension operation",
	)
	if err != nil {
		return verifiedPostgresOperations{}, err
	}
	extension, err := attestation.VerifyExtensionOperationArtifact(extensionDocument.bytes)
	if err != nil {
		return verifiedPostgresOperations{}, fmt.Errorf("extension operation verification failed")
	}
	cleanupDocument, err := readArtifact(
		settings.CleanupOperationPath().Value(),
		postgresv1.MaximumCredentialCleanupOperationBytes,
		"credential cleanup operation",
	)
	if err != nil {
		return verifiedPostgresOperations{}, err
	}
	cleanup, err := attestation.VerifyCredentialCleanupOperationArtifact(cleanupDocument.bytes)
	if err != nil {
		return verifiedPostgresOperations{}, fmt.Errorf("credential cleanup operation verification failed")
	}
	return verifiedPostgresOperations{
		bootstrap: bootstrap, extension: extension, cleanup: cleanup,
	}, nil
}
