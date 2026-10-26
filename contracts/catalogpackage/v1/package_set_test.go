package catalogpackagev1

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func TestDecodeCanonicalRequiresRealNonEmptyPayloads(t *testing.T) {
	set := validReleaseSet(t)
	document := canonicalSet(t, set)
	decoded, err := DecodeCanonical(document)
	if err != nil {
		t.Fatalf("decode valid package set: %v", err)
	}
	if len(decoded.Packages) != 1 || len(decoded.Packages[0].OwnerPayloads) != 1 {
		t.Fatalf("decoded package set = %#v", decoded)
	}

	tests := []struct {
		name   string
		mutate func(*ReleaseSet)
	}{
		{"empty packages", func(value *ReleaseSet) { value.Packages = nil }},
		{"empty owner payloads", func(value *ReleaseSet) {
			value.Packages[0].OwnerPayloads = nil
		}},
		{"fake payload hash", func(value *ReleaseSet) {
			value.Packages[0].OwnerPayloads[0].ContentSHA256 = strings.Repeat("0", 64)
		}},
		{"fake manifest hash", func(value *ReleaseSet) {
			value.Packages[0].ManifestSHA256 = strings.Repeat("0", 64)
		}},
		{"noncanonical signature", func(value *ReleaseSet) {
			value.Packages[0].Signature += "="
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := set.Clone()
			test.mutate(&candidate)
			if _, err := DecodeCanonical(canonicalSet(t, candidate)); err == nil {
				t.Fatal("invalid package set was accepted")
			}
		})
	}
}

func TestDecodeCanonicalRejectsNonCanonicalAndDuplicateComponent(t *testing.T) {
	set := validReleaseSet(t)
	canonical := canonicalSet(t, set)
	noncanonical := append([]byte("{ \n"), canonical[1:]...)
	if _, err := DecodeCanonical(noncanonical); err == nil {
		t.Fatal("noncanonical package set was accepted")
	}

	second := set.Packages[0]
	second.Manifest = json.RawMessage(`{"contract":"catalog-package.v2"}`)
	second.ManifestSHA256, _ = canonicaljsonv1.DigestHex(
		second.Manifest, MaximumManifestBytes,
	)
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	second.Signature = base64.RawURLEncoding.EncodeToString(
		ed25519.Sign(privateKey, second.Manifest),
	)
	set.Packages = append(set.Packages, second)
	if _, err := DecodeCanonical(canonicalSet(t, set)); err == nil {
		t.Fatal("duplicate component ID was accepted")
	}
}

func TestDecodeCanonicalRejectsNonDeterministicArrayOrdering(t *testing.T) {
	set := validReleaseSet(t)
	payload := set.Packages[0].OwnerPayloads[0]
	laterPayload := payload
	laterPayload.ComponentID = "e4000000-0000-0000-0000-000000000202"
	set.Packages[0].OwnerPayloads = []OwnerPayload{laterPayload, payload}
	if _, err := DecodeCanonical(canonicalSet(t, set)); err == nil {
		t.Fatal("descending owner payload order was accepted")
	}

	set = validReleaseSet(t)
	lower := set.Packages[0].clone()
	lower.Manifest = json.RawMessage(
		`{"contract":"catalog-package.v1","package_id":"e3000000-0000-0000-0000-000000000001","version":"1.0.0"}`,
	)
	lower.OwnerPayloads[0].ComponentID =
		"e4000000-0000-0000-0000-000000000200"
	resignPackage(t, &lower)
	set.Packages = append(set.Packages, lower)
	if _, err := DecodeCanonical(canonicalSet(t, set)); err == nil {
		t.Fatal("descending package identity order was accepted")
	}
}

func validReleaseSet(t *testing.T) ReleaseSet {
	t.Helper()
	manifest := json.RawMessage(
		`{"contract":"catalog-package.v1","package_id":"e3000000-0000-0000-0000-000000000002","version":"1.0.0"}`,
	)
	payload := json.RawMessage(`{"definition":{"fields":[{"name":"department_code"}]},"slug":"hr-department"}`)
	manifestDigest, err := canonicaljsonv1.DigestHex(manifest, MaximumManifestBytes)
	if err != nil {
		t.Fatal(err)
	}
	payloadDigest, err := canonicaljsonv1.DigestHex(payload, MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	return ReleaseSet{
		ContractVersion: ContractVersion,
		ReleaseID:       "018f0000-0000-7000-8000-000000000001",
		Packages: []SealedPackage{{
			Manifest: manifest, ManifestSHA256: manifestDigest,
			Signature: base64.RawURLEncoding.EncodeToString(
				ed25519.Sign(privateKey, manifest),
			),
			KeyID: "catalog-test-v1",
			OwnerPayloads: []OwnerPayload{{
				ComponentID: "e4000000-0000-0000-0000-000000000201",
				Owner:       "forms", Contract: "forms.catalog-definition.v1",
				ContentSHA256: payloadDigest, Document: payload,
			}},
		}},
	}
}

func resignPackage(t *testing.T, sealed *SealedPackage) {
	t.Helper()
	digest, err := canonicaljsonv1.DigestHex(
		sealed.Manifest, MaximumManifestBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	sealed.ManifestSHA256 = digest
	sealed.Signature = base64.RawURLEncoding.EncodeToString(
		ed25519.Sign(privateKey, sealed.Manifest),
	)
}

func canonicalSet(t *testing.T, set ReleaseSet) []byte {
	t.Helper()
	document, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumPackageSetBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
