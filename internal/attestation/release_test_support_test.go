package attestation

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

var releaseTestNow = time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)

func releaseTestDocuments(t *testing.T) ([]byte, []byte, TrustRoot) {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	rootDocument := mustReleaseJSON(t, map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "release-1", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
			"not_before": "2026-07-13T00:00:00Z", "not_after": "2026-07-14T00:00:00Z",
		}},
	})
	root, err := ParseTrustRoot(rootDocument)
	if err != nil {
		t.Fatalf("parse test root: %v", err)
	}
	manifest := mustReleaseJSON(t, validReleaseTestManifest())
	signature := base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, mustCanonical(t, manifest)))
	digest, err := DigestHex(manifest)
	if err != nil {
		t.Fatalf("digest manifest: %v", err)
	}
	signed := mustReleaseJSON(t, map[string]any{
		"contract_version": "release-signature.v1", "manifest_sha256": digest,
		"signature": map[string]any{"algorithm": "Ed25519", "key_id": "release-1", "value": signature},
	})
	return manifest, signed, root
}

func validReleaseTestManifest() map[string]any {
	digest := strings.Repeat("a", 64)
	return map[string]any{
		"contract_version": "postgres-release-manifest.v1",
		"release_id":       "44444444-4444-4444-8444-444444444444",
		"issued_at":        "2026-07-13T11:59:00Z", "expires_at": "2026-07-13T12:10:00Z",
		"database_name": "harpy", "database_origin": "v1_adopted",
		"database_encoding":        runtimev1.PostgresEncoding,
		"database_locale_provider": runtimev1.PostgresLocaleProvider,
		"database_collation":       runtimev1.PostgresCollation,
		"database_ctype":           runtimev1.PostgresCharacterType,
		"database_owner":           runtimev1.DatabaseOwnerRole,
		"data_checksums":           runtimev1.PostgresDataChecksums,
		"postgres_image_digest":    runtimev1.PostgresImageDigest,
		"postgres_engine_version":  runtimev1.PostgresEngineVersion,
		"vector_extension_version": runtimev1.VectorExtensionVersion,
		"runtime_generation":       2, "minimum_database_generation": 1,
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

func mustReleaseJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal release document: %v", err)
	}
	return document
}

func mustCanonical(t *testing.T, document []byte) []byte {
	t.Helper()
	canonical, err := Canonicalize(document)
	if err != nil {
		t.Fatalf("canonicalize release document: %v", err)
	}
	return canonical
}
