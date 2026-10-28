package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/tools/internal/firstpartycatalog"
)

func TestBuildCatalogPackageSetSignsAllFirstPartyPackages(t *testing.T) {
	authority := testCatalogPackageAuthority()
	releaseID := "018f0000-0000-7000-8000-000000000901"
	artifact, err := buildCatalogPackageSet(releaseID, authority)
	if err != nil {
		t.Fatal(err)
	}
	set, err := catalogpackagev1.DecodeCanonical(artifact.document)
	if err != nil {
		t.Fatal(err)
	}
	if set.ReleaseID != releaseID || len(set.Packages) != 9 {
		t.Fatalf("Catalog package set = %#v", set)
	}
	references, err := firstpartycatalog.BuildAll()
	if err != nil {
		t.Fatal(err)
	}
	byDigest := make(map[string]firstpartycatalog.Package, len(references))
	for _, reference := range references {
		byDigest[reference.ManifestSHA256] = reference
	}
	totalPayloads := 0
	for _, sealed := range set.Packages {
		reference, found := byDigest[sealed.ManifestSHA256]
		signature, decodeErr := base64.RawURLEncoding.DecodeString(sealed.Signature)
		if !found || decodeErr != nil ||
			string(sealed.Manifest) != string(reference.Manifest) ||
			!ed25519.Verify(
				authority.keys[sealed.KeyID], sealed.Manifest, signature,
			) {
			t.Fatalf(
				"first-party package signature binding failed: found=%t err=%v",
				found, decodeErr,
			)
		}
		if len(sealed.OwnerPayloads) != len(reference.OwnerPayloads) {
			t.Fatalf(
				"%s payload count = %d, want %d",
				reference.Slug, len(sealed.OwnerPayloads), len(reference.OwnerPayloads),
			)
		}
		totalPayloads += len(sealed.OwnerPayloads)
		previous := ""
		for _, payload := range sealed.OwnerPayloads {
			if previous >= payload.ComponentID {
				t.Fatalf(
					"%s payload order %q then %q",
					reference.Slug, previous, payload.ComponentID,
				)
			}
			previous = payload.ComponentID
		}
		delete(byDigest, sealed.ManifestSHA256)
	}
	if totalPayloads != 466 || len(byDigest) != 0 {
		t.Fatalf(
			"first-party package coverage = %d payloads, %d missing packages",
			totalPayloads, len(byDigest),
		)
	}
}

func TestCatalogPackageOrderUsesNumericSemver(t *testing.T) {
	base := firstpartycatalog.Package{
		PackageID: "018f0000-0000-7000-8000-000000000901",
	}
	versionTwo, versionTen := base, base
	versionTwo.Version = "2.0.0"
	versionTen.Version = "10.0.0"
	if !catalogPackageLess(versionTwo, versionTen) ||
		catalogPackageLess(versionTen, versionTwo) {
		t.Fatal("Catalog package ordering compared semantic versions lexically")
	}
}
