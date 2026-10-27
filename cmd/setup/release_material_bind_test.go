package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestReleaseMaterialBindsEmbeddedSchemaContract(t *testing.T) {
	lines, err := releaseMaterial(t.TempDir(), "single-tenant")
	if err != nil {
		t.Fatalf("release material: %v", err)
	}
	paths := releasePaths(t, lines)
	embedded, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := attestation.VerifyBinarySchemaContract(embedded)
	if err != nil {
		t.Fatal(err)
	}
	expected := contract.Value()
	migrationBytes := readReleaseFile(t, paths["HARPY_MIGRATION_MANIFEST_PATH"])
	extensionBytes := readReleaseFile(t, paths["HARPY_EXTENSION_MANIFEST_PATH"])
	if digestHex(migrationBytes) != expected.MigrationManifestSHA256 ||
		digestHex(extensionBytes) != expected.ExtensionManifestSHA256 {
		t.Fatalf("written digests migration %s extension %s",
			digestHex(migrationBytes), digestHex(extensionBytes))
	}
	extension, err := attestation.VerifyExtensionManifestArtifact(extensionBytes)
	if err != nil {
		t.Fatal(err)
	}
	trust, err := attestation.ParseTrustRoot(readReleaseFile(t, paths["HARPY_RELEASE_TRUST_ROOT_PATH"]))
	if err != nil {
		t.Fatal(err)
	}
	release, err := attestation.VerifyRelease(
		readReleaseFile(t, paths["HARPY_RELEASE_MANIFEST_PATH"]),
		readReleaseFile(t, paths["HARPY_RELEASE_SIGNATURE_PATH"]),
		trust, time.Now().UTC(), "harpy",
	)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := attestation.VerifyMigrationManifestArtifact(migrationBytes)
	if err != nil {
		t.Fatalf("schema contract binding rejected the embedded migration manifest: %v", err)
	}
	built, err := attestation.BuildBinarySchemaContract(release, migration, extension)
	if err != nil || !built.IsVerified() || !bytes.Equal(built.CanonicalJSON(), embedded) {
		t.Fatalf("schema contract binding: %v", err)
	}
}

func releasePaths(t *testing.T, lines []string) map[string]string {
	t.Helper()
	paths := make(map[string]string, len(lines))
	for _, line := range lines {
		key, path, found := strings.Cut(line, "=")
		if !found {
			t.Fatalf("malformed release line %s", line)
		}
		paths[key] = path
	}
	return paths
}

func readReleaseFile(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func digestHex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
