package schemacontract

import (
	"bytes"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestLoadBoundStartupMatchesContractDigest(t *testing.T) {
	contract, err := LoadBoundStartup()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		t.Fatal(err)
	}
	if !contract.IsVerified() || contract.Value().MigrationManifestSHA256 == "" ||
		contract.Value().ContractVersion != runtimev1.BinarySchemaContractVersion ||
		len(manifest) == 0 {
		t.Fatalf("bound startup = %+v manifest %d", contract.Value(), len(manifest))
	}
	if contract.Value().RuntimeGeneration != 1 ||
		contract.Value().MinimumDatabaseGeneration != 1 {
		t.Fatalf("generations = %+v", contract.Value())
	}
}

func TestLoadBoundStartupRejectsStaleAndInvalidManifests(t *testing.T) {
	manifest, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := binarycontract.Verify(contract)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = loadBoundStartup(
		func() ([]byte, error) { return []byte("{"), nil },
		func() (binarycontract.Artifact, error) { return artifact, nil },
	); err == nil {
		t.Fatal("invalid manifest was accepted")
	}
	wrong := contractArtifact(t, strings.Repeat("ab", 32), "extension-stale")
	if _, err = loadBoundStartup(
		func() ([]byte, error) { return manifest, nil },
		func() (binarycontract.Artifact, error) { return wrong, nil },
	); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("stale contract error = %v", err)
	}
}

func TestEmbeddedFreshManifestVerifies(t *testing.T) {
	manifest, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		t.Fatal(err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(manifest, verified.CanonicalJSON()) {
		t.Fatal("embedded fresh manifest is not canonical")
	}
	contract, err := LoadBoundStartup()
	if err != nil || contract.Value().MigrationManifestSHA256 != verified.DigestSHA256() {
		t.Fatalf("bound digest = %s, %v", verified.DigestSHA256(), err)
	}
}

func TestLoadBoundStartupRejectsNonCanonicalManifest(t *testing.T) {
	manifest, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := binarycontract.Verify(contract)
	if err != nil {
		t.Fatal(err)
	}
	pretty := bytes.Replace(manifest, []byte(`,"engine"`), []byte(`, "engine"`), 1)
	if bytes.Equal(pretty, manifest) {
		t.Fatal("manifest had no engine field")
	}
	_, err = loadBoundStartup(
		func() ([]byte, error) { return pretty, nil },
		func() (binarycontract.Artifact, error) { return artifact, nil },
	)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("non-canonical error = %v", err)
	}
}
