package main

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestCanonicalLocalManifestsValidate(t *testing.T) {
	t.Parallel()
	values := []interface{ Validate() error }{
		extensionManifest(), rolesManifest(), privilegesManifest(),
		ownershipManifest("10000000-0000-4000-8000-000000000001"),
	}
	for _, value := range values {
		if err := value.Validate(); err != nil {
			t.Fatalf("manifest validation failed: %v", err)
		}
	}
	if extensionManifest().PostgresImageDigest != postgresv1.PostgresImageDigest {
		t.Fatal("extension image digest differs from the PostgreSQL contract")
	}
}

func TestExtensionControlEvidenceIsExact(t *testing.T) {
	t.Parallel()
	manifest := extensionManifest()
	if len(manifest.Extensions) != 2 {
		t.Fatalf("extension count = %d", len(manifest.Extensions))
	}
	if manifest.Extensions[0].ControlFileSHA256 != pgTrgmControlHash ||
		manifest.Extensions[1].ControlFileSHA256 != vectorControlHash {
		t.Fatal("extension control evidence differs")
	}
	value, err := newArtifact("extension-manifest.json", manifest)
	if err != nil {
		t.Fatal(err)
	}
	if value.digest != extensionManifestHash {
		t.Fatalf("extension manifest sha256 = %s", value.digest)
	}
}

func TestOperationContractsBindManifestDigests(t *testing.T) {
	t.Parallel()
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	operations := []interface{ Validate() error }{
		bootstrapOperation(digest, digest, digest),
		extensionOperation(digest, digest),
		cleanupOperation(digest, digest),
	}
	for _, operation := range operations {
		if err := operation.Validate(); err != nil {
			t.Fatalf("operation validation failed: %v", err)
		}
	}
}
