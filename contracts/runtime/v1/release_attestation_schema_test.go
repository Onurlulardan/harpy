package runtimev1

import (
	"strings"
	"testing"
)

func TestReleaseAttestationSchemasAcceptExactContracts(t *testing.T) {
	digest := strings.Repeat("a", 64)
	publicKey := strings.Repeat("A", 43)
	tests := []struct {
		schema string
		value  map[string]any
	}{
		{"release-manifest.schema.json", validReleaseManifest(digest)},
		{"release-signature.schema.json", map[string]any{
			"contract_version": "release-signature.v1", "manifest_sha256": digest,
			"signature": validReceiptSignature(),
		}},
		{"attestation-trust-root.schema.json", map[string]any{
			"contract_version": "attestation-trust-root.v1",
			"keys": []any{map[string]any{
				"key_id": "release-1", "algorithm": "Ed25519", "public_key": publicKey,
				"not_before": "2026-07-13T00:00:00Z", "not_after": "2026-07-14T00:00:00Z",
			}},
		}},
		{"attestation-signing-key.schema.json", map[string]any{
			"contract_version": "attestation-signing-key.v1", "key_id": "release-1",
			"algorithm": "Ed25519", "seed": publicKey,
		}},
	}
	for _, test := range tests {
		t.Run(test.schema, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.value); err != nil {
				t.Fatalf("valid attestation contract rejected: %v", err)
			}
		})
	}
}

func TestReleaseAttestationSchemasRejectDrift(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		schema string
		value  map[string]any
	}{
		{"wrong image", "release-manifest.schema.json", withReleaseValue(digest, "postgres_image_digest", "sha256:"+strings.Repeat("b", 64))},
		{"wrong owner", "release-manifest.schema.json", withReleaseValue(digest, "database_owner", "postgres")},
		{"wrong locale provider", "release-manifest.schema.json", withReleaseValue(digest, "database_locale_provider", "icu")},
		{"checksum disabled", "release-manifest.schema.json", withReleaseValue(digest, "data_checksums", false)},
		{"unknown field", "release-manifest.schema.json", withReleaseValue(digest, "temporary", true)},
		{"weak signature", "release-signature.schema.json", map[string]any{
			"contract_version": "release-signature.v1", "manifest_sha256": digest,
			"signature": map[string]any{"algorithm": "HS256", "key_id": "release-1", "value": strings.Repeat("A", 86)},
		}},
		{"padded public key", "attestation-trust-root.schema.json", map[string]any{
			"contract_version": "attestation-trust-root.v1", "keys": []any{map[string]any{
				"key_id": "release-1", "algorithm": "Ed25519", "public_key": strings.Repeat("A", 43) + "=",
				"not_before": "2026-07-13T00:00:00Z", "not_after": "2026-07-14T00:00:00Z",
			}},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.value); err == nil {
				t.Fatal("unsafe attestation contract accepted")
			}
		})
	}
}

func validReleaseManifest(digest string) map[string]any {
	return map[string]any{
		"contract_version": "postgres-release-manifest.v1",
		"release_id":       "44444444-4444-4444-8444-444444444444",
		"issued_at":        "2026-07-13T00:00:00Z", "expires_at": "2026-07-14T00:00:00Z",
		"database_name": "harpy", "database_origin": "v1_adopted",
		"database_owner": "harpy_owner", "database_encoding": "UTF8",
		"database_locale_provider": "libc", "database_collation": "en_US.utf8",
		"database_ctype": "en_US.utf8", "data_checksums": true,
		"postgres_image_digest":   PostgresImageDigest,
		"postgres_engine_version": "18.4", "vector_extension_version": "0.8.3",
		"runtime_generation": 1, "minimum_database_generation": 1,
		"migration_manifest_sha256": digest, "extension_manifest_sha256": digest,
		"roles_manifest_sha256": digest, "privileges_manifest_sha256": digest,
		"ownership_manifest_sha256":           digest,
		"bootstrap_authority_manifest_sha256": digest,
		"bootstrap_operation_digest":          digest, "extension_operation_digest": digest,
		"credential_cleanup_operation_digest":         digest,
		"cutover_authority_adoption_operation_digest": digest,
		"catalog_package_set_sha256":                  digest,
	}
}

func withReleaseValue(digest, key string, value any) map[string]any {
	manifest := validReleaseManifest(digest)
	manifest[key] = value
	return manifest
}
