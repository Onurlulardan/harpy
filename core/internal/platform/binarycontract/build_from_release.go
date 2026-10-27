package binarycontract

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func Build(
	release attestation.VerifiedRelease,
	migration attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	extension attestation.VerifiedPostgresArtifact[postgresv1.ExtensionManifest],
) (Artifact, error) {
	verified, err := attestation.BuildBinarySchemaContract(release, migration, extension)
	if err != nil {
		return Artifact{}, err
	}
	return wrap(verified), nil
}
