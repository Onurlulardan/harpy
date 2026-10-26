package catalog

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const catalogFinalizerReleaseID = "018f0000-0000-7000-8000-000000000401"
const catalogFinalizerPackageID = "018f0000-0000-7000-8000-000000000402"
const catalogFinalizerComponentID = "018f0000-0000-7000-8000-000000000403"

var catalogFinalizerRegisteredAt = time.Date(
	2026, 7, 30, 10, 0, 0, 123456000, time.UTC,
)

func catalogFinalizerEvidence(
	t *testing.T,
) (
	attestation.VerifiedRelease,
	packageregistry.SealedVersion,
	ed25519.PublicKey,
) {
	t.Helper()
	releasePrivateKey := ed25519.NewKeyFromSeed(
		bytes.Repeat([]byte{7}, ed25519.SeedSize),
	)
	releasePublicKey := releasePrivateKey.Public().(ed25519.PublicKey)
	rootDocument := catalogFinalizerJSON(t, map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "release-2026-07", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(releasePublicKey),
			"not_before": "2026-07-30T09:00:00Z",
			"not_after":  "2026-07-30T11:00:00Z",
		}},
	})
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		t.Fatal(err)
	}
	releaseDocument := catalogFinalizerReleaseDocument(t)
	releaseCanonical, err := attestation.Canonicalize(releaseDocument)
	if err != nil {
		t.Fatal(err)
	}
	releaseDigest, err := attestation.DigestHex(releaseDocument)
	if err != nil {
		t.Fatal(err)
	}
	signatureDocument := catalogFinalizerJSON(t, map[string]any{
		"contract_version": "release-signature.v1",
		"manifest_sha256":  releaseDigest,
		"signature": map[string]any{
			"algorithm": "Ed25519", "key_id": "release-2026-07",
			"value": base64.RawURLEncoding.EncodeToString(
				ed25519.Sign(releasePrivateKey, releaseCanonical),
			),
		},
	})
	release, err := attestation.VerifyRelease(
		releaseDocument, signatureDocument, root,
		time.Date(2026, 7, 30, 10, 1, 0, 0, time.UTC), "harpy",
	)
	if err != nil {
		t.Fatal(err)
	}
	packagePrivateKey := ed25519.NewKeyFromSeed(
		bytes.Repeat([]byte{8}, ed25519.SeedSize),
	)
	return release, catalogFinalizerPackage(t, packagePrivateKey),
		packagePrivateKey.Public().(ed25519.PublicKey)
}

func catalogFinalizerReleaseDocument(t *testing.T) []byte {
	t.Helper()
	digest := strings.Repeat("a", 64)
	return catalogFinalizerJSON(t, map[string]any{
		"contract_version": "postgres-release-manifest.v1",
		"release_id":       catalogFinalizerReleaseID,
		"issued_at":        "2026-07-30T10:00:00.123456789Z",
		"expires_at":       "2026-07-30T10:10:00Z",
		"database_name":    "harpy", "database_origin": "fresh",
		"database_encoding":        runtimev1.PostgresEncoding,
		"database_locale_provider": runtimev1.PostgresLocaleProvider,
		"database_collation":       runtimev1.PostgresCollation,
		"database_ctype":           runtimev1.PostgresCharacterType,
		"database_owner":           runtimev1.DatabaseOwnerRole,
		"data_checksums":           runtimev1.PostgresDataChecksums,
		"postgres_image_digest":    runtimev1.PostgresImageDigest,
		"postgres_engine_version":  runtimev1.PostgresEngineVersion,
		"vector_extension_version": runtimev1.VectorExtensionVersion,
		"runtime_generation":       1, "minimum_database_generation": 1,
		"migration_manifest_sha256": digest, "extension_manifest_sha256": digest,
		"roles_manifest_sha256": digest, "privileges_manifest_sha256": digest,
		"ownership_manifest_sha256":           digest,
		"bootstrap_authority_manifest_sha256": digest,
		"bootstrap_operation_digest":          digest, "extension_operation_digest": digest,
		"credential_cleanup_operation_digest":         digest,
		"cutover_authority_adoption_operation_digest": digest,
		"catalog_package_set_sha256":                  digest,
	})
}

func catalogFinalizerPackage(
	t *testing.T,
	privateKey ed25519.PrivateKey,
) packageregistry.SealedVersion {
	t.Helper()
	payload, err := canonicaljsonv1.CanonicalizeObject(
		[]byte(`{"component_id":"`+catalogFinalizerComponentID+
			`","definition":{"slug":"release-proof"}}`),
		catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	payloadDigest := sha256.Sum256(payload)
	payloadDigestHex := hex.EncodeToString(payloadDigest[:])
	document := []byte(`{"contract":"catalog-package.v1","package_id":"` +
		catalogFinalizerPackageID + `","slug":"release-proof","version":"1.0.0",` +
		`"platform_contract":"harpy.v2","profile":"core","dependencies":[],` +
		`"components":[{"ID":"` + catalogFinalizerComponentID + `","Owner":"forms",` +
		`"Contract":"forms.v2","kind":"FORM","content_sha256":"` +
		payloadDigestHex + `"}],"menu":[]}`)
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, 1<<20,
	)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	return packageregistry.SealedVersion{
		PackageID: catalogFinalizerPackageID, Slug: "release-proof", Version: "1.0.0",
		Manifest: canonical, ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature: ed25519.Sign(privateKey, canonical), KeyID: "catalog-2026-07",
		ReleaseID:    "018f0000-0000-7000-8000-000000000499",
		RegisteredAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		OwnerPayloads: []catalogpackagev1.OwnerPayload{{
			ComponentID: catalogFinalizerComponentID, Owner: "forms",
			Contract: "forms.v2", ContentSHA256: payloadDigestHex,
			Document: payload,
		}},
	}
}

func catalogFinalizerJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
