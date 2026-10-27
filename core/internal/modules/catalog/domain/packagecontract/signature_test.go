package packagecontract

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
)

const (
	packageA   = "018f0000-0000-7000-8000-000000000001"
	packageB   = "018f0000-0000-7000-8000-000000000002"
	componentA = "018f0000-0000-7000-8000-000000000101"
)

func TestVerifyRejectsManifestAndHashTampering(t *testing.T) {
	document := manifestDocument(packageA, "[]")
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	attestation := Attestation{ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature: ed25519.Sign(privateKey, document), KeyID: "catalog-test-v1"}
	if _, err = Verify(document, attestation, publicKey); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	tampered := append([]byte(nil), document...)
	tampered[len(tampered)-2] ^= 1
	if _, err = Verify(tampered, attestation, publicKey); !errors.Is(err, ErrManifestTampered) {
		t.Fatalf("tampered document error = %v", err)
	}
	attestation.ManifestSHA256 = fmt.Sprintf("%064d", 0)
	if _, err = Verify(document, attestation, publicKey); !errors.Is(err, ErrManifestTampered) {
		t.Fatalf("tampered hash error = %v", err)
	}
}

func TestResolveDependenciesRejectsCycle(t *testing.T) {
	root := decodeManifest(t, manifestDocument(packageA, dependencyJSON(packageB)))
	dependency := decodeManifest(t, manifestDocument(packageB, dependencyJSON(packageA)))
	_, err := ResolveDependencies(root, map[string][]Manifest{packageB: {dependency}, packageA: {root}})
	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("cycle error = %v", err)
	}
}

func decodeManifest(t *testing.T, document []byte) Manifest {
	t.Helper()
	value, err := Decode(document)
	if err != nil {
		t.Fatalf("decode manifest: %v\n%s", err, document)
	}
	return value
}

func dependencyJSON(packageID string) string {
	return fmt.Sprintf(`[{"package_id":"%s","range":"^1.0.0","kind":"required"}]`, packageID)
}

func manifestDocument(packageID, dependencies string) []byte {
	return []byte(fmt.Sprintf(`{"contract":"catalog-package.v1","package_id":"%s","slug":"package-test","version":"1.0.0","platform_contract":"harpy.v2","profile":"business","dependencies":%s,"components":[{"ID":"%s","Owner":"forms","Contract":"forms.v2","kind":"FORM","content_sha256":"%064d"}],"menu":[]}`, packageID, dependencies, componentA, 1))
}
