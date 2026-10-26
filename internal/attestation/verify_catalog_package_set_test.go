package attestation

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

type catalogPackageTestKeyring struct {
	id  string
	key ed25519.PublicKey
}

func (keyring catalogPackageTestKeyring) Resolve(
	keyID string,
) (ed25519.PublicKey, error) {
	if keyID != keyring.id {
		return nil, errors.New("key unavailable")
	}
	return append(ed25519.PublicKey(nil), keyring.key...), nil
}

func (keyring catalogPackageTestKeyring) PublicKeys() []ed25519.PublicKey {
	return []ed25519.PublicKey{append(ed25519.PublicKey(nil), keyring.key...)}
}

func TestVerifyCatalogPackageSetRequiresDistinctPackageAuthority(t *testing.T) {
	releaseSeed := make([]byte, ed25519.SeedSize)
	packageSeed := make([]byte, ed25519.SeedSize)
	for index := range releaseSeed {
		releaseSeed[index] = byte(index + 1)
		packageSeed[index] = byte(index + 65)
	}
	releasePrivate := ed25519.NewKeyFromSeed(releaseSeed)
	packagePrivate := ed25519.NewKeyFromSeed(packageSeed)
	setDocument := catalogPackageSetDocument(t, packagePrivate)
	setDigest := sha256.Sum256(setDocument)
	manifest := validReleaseTestManifest()
	manifest["catalog_package_set_sha256"] = hex.EncodeToString(setDigest[:])
	manifestDocument := mustReleaseJSON(t, manifest)
	rootDocument := mustReleaseJSON(t, map[string]any{
		"contract_version": "attestation-trust-root.v1",
		"keys": []any{map[string]any{
			"key_id": "release-1", "algorithm": "Ed25519",
			"public_key": base64.RawURLEncoding.EncodeToString(
				releasePrivate.Public().(ed25519.PublicKey),
			),
			"not_before": "2026-07-13T00:00:00Z",
			"not_after":  "2026-07-14T00:00:00Z",
		}},
	})
	root, err := ParseTrustRoot(rootDocument)
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest, err := DigestHex(manifestDocument)
	if err != nil {
		t.Fatal(err)
	}
	signatureDocument := mustReleaseJSON(t, map[string]any{
		"contract_version": "release-signature.v1",
		"manifest_sha256":  manifestDigest,
		"signature": map[string]any{
			"algorithm": "Ed25519", "key_id": "release-1",
			"value": base64.RawURLEncoding.EncodeToString(
				ed25519.Sign(releasePrivate, manifestDocument),
			),
		},
	})
	release, err := VerifyRelease(
		manifestDocument, signatureDocument, root, releaseTestNow, "harpy",
	)
	if err != nil {
		t.Fatal(err)
	}
	keyring := catalogPackageTestKeyring{
		id: "catalog-1", key: packagePrivate.Public().(ed25519.PublicKey),
	}
	verified, err := VerifyCatalogPackageSet(
		setDocument, release, keyring, root.ActivePublicKeys(release.IssuedAt()),
	)
	if err != nil || !verified.IsVerified() || len(verified.Value().Packages) != 1 {
		t.Fatalf("verify Catalog package set: %#v, %v", verified, err)
	}
	shared := catalogPackageTestKeyring{
		id: "catalog-1", key: releasePrivate.Public().(ed25519.PublicKey),
	}
	if _, err = VerifyCatalogPackageSet(
		setDocument, release, shared, root.ActivePublicKeys(release.IssuedAt()),
	); err == nil {
		t.Fatal("release signing authority was accepted as Catalog authority")
	}
}

func catalogPackageSetDocument(
	t *testing.T,
	privateKey ed25519.PrivateKey,
) []byte {
	t.Helper()
	manifest := mustReleaseJSON(t, map[string]any{
		"package_id": "018f0000-0000-7000-8000-000000000010",
		"version":    "1.0.0",
	})
	manifestDigest := sha256.Sum256(manifest)
	payloadDocument := mustReleaseJSON(t, map[string]any{"definition": "test"})
	payloadDigest := sha256.Sum256(payloadDocument)
	set := catalogpackagev1.ReleaseSet{
		ContractVersion: catalogpackagev1.ContractVersion,
		ReleaseID:       "44444444-4444-4444-8444-444444444444",
		Packages: []catalogpackagev1.SealedPackage{{
			Manifest:       manifest,
			ManifestSHA256: hex.EncodeToString(manifestDigest[:]),
			Signature: base64.RawURLEncoding.EncodeToString(
				ed25519.Sign(privateKey, manifest),
			),
			KeyID: "catalog-1",
			OwnerPayloads: []catalogpackagev1.OwnerPayload{{
				ComponentID: "018f0000-0000-7000-8000-000000000001",
				Owner:       "forms", Contract: "forms.catalog-definition.v1",
				ContentSHA256: hex.EncodeToString(payloadDigest[:]),
				Document:      payloadDocument,
			}},
		}},
	}
	document, err := MarshalCanonical(set)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
