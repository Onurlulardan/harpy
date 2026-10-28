package migrationbundle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestCanonicalFreshMatchesBinarySchemaContract(t *testing.T) {
	root := moduleRoot(t)
	body, err := CanonicalFresh(root)
	if err != nil {
		t.Fatal(err)
	}
	again, err := CanonicalFresh(root)
	if err != nil {
		t.Fatal(err)
	}
	if Digest(body) != Digest(again) {
		t.Fatal("fresh manifest digest is not stable")
	}
	document, err := os.ReadFile(filepath.Join(
		root, "contracts", "runtime", "v1", "binary-schema-contract.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	contract, err := attestation.VerifyBinarySchemaContract(document)
	if err != nil {
		t.Fatal(err)
	}
	if Digest(body) != contract.Value().MigrationManifestSHA256 {
		t.Fatalf("fresh digest %s contract %s", Digest(body), contract.Value().MigrationManifestSHA256)
	}
	artifact, err := attestation.VerifyMigrationManifestArtifact(body)
	if err != nil {
		t.Fatal(err)
	}
	if artifact.DigestSHA256() != Digest(body) {
		t.Fatalf("verifier digest %s", artifact.DigestSHA256())
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root is unavailable")
		}
		dir = parent
	}
}
