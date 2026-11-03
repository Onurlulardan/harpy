package migrator

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestTenantMigrationReleaseMustMatchEmbeddedBinaryContract(t *testing.T) {
	digest := strings.Repeat("a", 64)
	binary := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component: "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
	}
	manifest := attestation.ReleaseManifest{
		ReleaseID: "11111111-1111-4111-8111-111111111111",
		RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: digest, ExtensionManifestSHA256: digest,
		PostgresImageDigest: runtimev1.PostgresImageDigest,
		PostgresEngineVersion: runtimev1.PostgresEngineVersion,
		VectorExtensionVersion: runtimev1.VectorExtensionVersion,
	}
	if err := validateTenantMigrationReleaseContract(manifest, binary); err != nil {
		t.Fatalf("matching signed release was rejected: %v", err)
	}
	mutations := []struct {
		name string
		edit func(*attestation.ReleaseManifest)
	}{
		{"runtime generation", func(m *attestation.ReleaseManifest) { m.RuntimeGeneration++ }},
		{"database generation", func(m *attestation.ReleaseManifest) { m.MinimumDatabaseGeneration++ }},
		{"migration manifest", func(m *attestation.ReleaseManifest) { m.MigrationManifestSHA256 = strings.Repeat("b", 64) }},
		{"extension manifest", func(m *attestation.ReleaseManifest) { m.ExtensionManifestSHA256 = strings.Repeat("b", 64) }},
		{"PostgreSQL image", func(m *attestation.ReleaseManifest) { m.PostgresImageDigest = "sha256:" + strings.Repeat("b", 64) }},
		{"PostgreSQL engine", func(m *attestation.ReleaseManifest) { m.PostgresEngineVersion = "17.0" }},
		{"vector extension", func(m *attestation.ReleaseManifest) { m.VectorExtensionVersion = "0.0.0" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := manifest
			mutation.edit(&changed)
			if err := validateTenantMigrationReleaseContract(changed, binary); err == nil {
				t.Fatal("mismatched signed release was accepted")
			}
		})
	}
}
