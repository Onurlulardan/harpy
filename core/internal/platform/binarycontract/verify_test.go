package binarycontract

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestVerifyReturnsOpaqueCanonicalArtifact(t *testing.T) {
	digest := strings.Repeat("a", 64)
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
	}
	document, err := attestation.MarshalCanonical(contract)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := Verify(document)
	if err != nil {
		t.Fatalf("verify contract: %v", err)
	}
	if !artifact.IsVerified() || artifact.Value() != contract || artifact.DigestSHA256() == "" {
		t.Fatalf("unexpected verified artifact: %+v", artifact.Value())
	}
}
