package schemacontract

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func contractArtifact(
	t *testing.T,
	migrationDigest string,
	extensionSeed string,
) binarycontract.Artifact {
	t.Helper()
	extensionDigest := sha256.Sum256([]byte(extensionSeed))
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: migrationDigest,
		ExtensionManifestSHA256: hex.EncodeToString(extensionDigest[:]),
		PostgresImageDigest:     runtimev1.PostgresImageDigest,
	}
	document, err := attestation.MarshalCanonical(contract)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := binarycontract.Verify(document)
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}
