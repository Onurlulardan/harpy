package attestation

import (
	"encoding/json"
	"testing"
)

func TestVerifyPostgresArtifactUsesStrictJSONAndJCSDigest(t *testing.T) {
	manifest := testMigrationManifest()
	structDocument := testDocument(t, manifest)
	var reordered map[string]any
	if err := json.Unmarshal(structDocument, &reordered); err != nil {
		t.Fatal(err)
	}
	mapDocument := testDocument(t, reordered)
	first, err := VerifyMigrationManifestArtifact(structDocument)
	if err != nil {
		t.Fatal(err)
	}
	second, err := VerifyMigrationManifestArtifact(mapDocument)
	if err != nil {
		t.Fatal(err)
	}
	if first.DigestSHA256() != second.DigestSHA256() ||
		string(first.CanonicalJSON()) != string(second.CanonicalJSON()) {
		t.Fatal("equivalent property order changed canonical artifact identity")
	}
	expectedDigest, err := DigestHex(structDocument)
	if err != nil || first.DigestSHA256() != expectedDigest {
		t.Fatalf("artifact digest does not equal SHA256(JCS(document)): %v", err)
	}
	value := first.Value()
	value.Entries[0].Owner = "mutated"
	if first.Value().Entries[0].Owner == "mutated" {
		t.Fatal("verified artifact Value aliases internal state")
	}
}

func TestVerifyPostgresArtifactAllowsAbsentOriginVariant(t *testing.T) {
	manifest := testMigrationManifest()
	manifest.Entries[0].Variants.V1Adopted = nil
	artifact, err := VerifyMigrationManifestArtifact(testDocument(t, manifest))
	if err != nil {
		t.Fatal(err)
	}
	selected, err := artifact.Value().Select("fresh")
	if err != nil || len(selected) != 1 {
		t.Fatalf("fresh-only migration could not be selected: %v", err)
	}
}

func TestVerifyPostgresArtifactRejectsStrictAndSemanticDrift(t *testing.T) {
	manifest := testMigrationManifest()
	base := testDocument(t, manifest)
	var unknown map[string]any
	if err := json.Unmarshal(base, &unknown); err != nil {
		t.Fatal(err)
	}
	unknown["temporary"] = true
	duplicate := append([]byte(`{"contract_version":"postgres-migration-manifest.v1",`), base[1:]...)
	nullVariant := cloneJSONMap(t, base)
	nullVariant["entries"].([]any)[0].(map[string]any)["variants"].(map[string]any)["fresh"] = nil
	invalidSemantic := manifest
	invalidSemantic.Entries[0].Variants.Fresh.Path =
		"migrations/20260714010000_foundation.v1_adopted.sql"
	tests := []struct {
		name     string
		document []byte
	}{
		{"unknown", testDocument(t, unknown)},
		{"duplicate", duplicate},
		{"null optional", testDocument(t, nullVariant)},
		{"semantic mismatch", testDocument(t, invalidSemantic)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := VerifyMigrationManifestArtifact(test.document); err == nil {
				t.Fatal("unsafe PostgreSQL artifact accepted")
			}
		})
	}
}

func cloneJSONMap(t *testing.T, document []byte) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(document, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
