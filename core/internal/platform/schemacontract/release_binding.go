package schemacontract

import (
	"bytes"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type ReleaseBinding struct {
	contract EmbeddedContract
	verified bool
}

func BindRelease(
	contract EmbeddedContract,
	release attestation.VerifiedRelease,
	migration attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	extension attestation.VerifiedPostgresArtifact[postgresv1.ExtensionManifest],
) (ReleaseBinding, error) {
	expected, err := binarycontract.Build(release, migration, extension)
	if err != nil {
		return ReleaseBinding{}, fmt.Errorf("build release binary contract: %w", err)
	}
	return bindExpected(contract, expected)
}

func bindExpected(
	contract EmbeddedContract,
	expected binarycontract.Artifact,
) (ReleaseBinding, error) {
	if !contract.IsVerified() || !expected.IsVerified() ||
		!bytes.Equal(contract.artifact.CanonicalJSON(), expected.CanonicalJSON()) {
		return ReleaseBinding{}, fmt.Errorf("embedded binary contract differs from verified release")
	}
	return ReleaseBinding{contract: contract, verified: true}, nil
}

func (binding ReleaseBinding) Contract() EmbeddedContract {
	if !binding.verified {
		return EmbeddedContract{}
	}
	return binding.contract
}

func (binding ReleaseBinding) IsVerified() bool {
	return binding.verified && binding.contract.IsVerified()
}
