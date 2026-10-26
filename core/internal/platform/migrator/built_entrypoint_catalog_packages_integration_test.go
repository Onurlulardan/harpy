//go:build integration

package migrator

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const builtEntrypointCatalogPackageKeyID = "built-entrypoint-catalog-v1"

type builtEntrypointReferencePackage struct {
	id     string
	sealed catalogpackagev1.SealedPackage
}

type builtEntrypointOwnerGroup struct {
	Contract string                          `json:"contract"`
	Owner    string                          `json:"owner"`
	Payloads []catalogpackagev1.OwnerPayload `json:"payloads"`
}

func prepareBuiltEntrypointCatalogPackageSet(
	t *testing.T,
	directory string,
	stage builtEntrypointStage,
) (builtEntrypointArtifact, string) {
	t.Helper()
	seed := sha256.Sum256([]byte(
		"harpy-built-entrypoint-catalog-package-authority.v1",
	))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	defer clear(privateKey)
	references := readBuiltEntrypointReferencePackages(t, stage, privateKey)
	set := catalogpackagev1.ReleaseSet{
		ContractVersion: catalogpackagev1.ContractVersion,
		ReleaseID:       builtEntrypointReleaseID,
		Packages:        make([]catalogpackagev1.SealedPackage, len(references)),
	}
	for index := range references {
		set.Packages[index] = references[index].sealed
	}
	raw, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("encode built entrypoint Catalog package set: %v", err)
	}
	document, err := canonicaljsonv1.CanonicalizeObject(
		raw, catalogpackagev1.MaximumPackageSetBytes,
	)
	if err != nil {
		t.Fatalf("canonicalize built entrypoint Catalog package set: %v", err)
	}
	if _, err := catalogpackagev1.DecodeCanonical(document); err != nil {
		t.Fatalf("validate built entrypoint Catalog package set: %v", err)
	}
	artifact := writeBuiltEntrypointCatalogPackageSet(
		t, directory, "catalog-package-set.json", document,
	)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	keyring, err := json.Marshal(map[string]any{
		"version": 1,
		"keys": []map[string]string{{
			"key_id":     builtEntrypointCatalogPackageKeyID,
			"public_key": base64.RawURLEncoding.EncodeToString(publicKey),
		}},
	})
	if err != nil {
		t.Fatalf("encode built entrypoint Catalog package keyring: %v", err)
	}
	return artifact, string(keyring)
}

func readBuiltEntrypointReferencePackages(
	t *testing.T,
	stage builtEntrypointStage,
	privateKey ed25519.PrivateKey,
) []builtEntrypointReferencePackage {
	t.Helper()
	root := filepath.Join(
		stage.directory, "tools", "internal", "firstpartycatalog", "testdata",
	)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read staged first-party Catalog packages: %v", err)
	}
	result := make([]builtEntrypointReferencePackage, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		result = append(result, readBuiltEntrypointReferencePackage(
			t, filepath.Join(root, entry.Name(), "artifacts"), privateKey,
		))
	}
	if len(result) != 9 {
		t.Fatalf("staged first-party Catalog package count = %d", len(result))
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].id < result[right].id
	})
	return result
}

func readBuiltEntrypointReferencePackage(
	t *testing.T,
	directory string,
	privateKey ed25519.PrivateKey,
) builtEntrypointReferencePackage {
	t.Helper()
	manifest := readBuiltEntrypointCanonicalJSON(
		t, filepath.Join(directory, "manifest.json"),
		catalogpackagev1.MaximumManifestBytes,
	)
	var identity struct {
		PackageID string `json:"package_id"`
	}
	if err := json.Unmarshal(manifest, &identity); err != nil || identity.PackageID == "" {
		t.Fatalf("decode first-party Catalog package identity in %s: %v", directory, err)
	}
	payloads := readBuiltEntrypointOwnerPayloads(t, directory)
	sort.Slice(payloads, func(left, right int) bool {
		return payloads[left].ComponentID < payloads[right].ComponentID
	})
	digest := sha256.Sum256(manifest)
	return builtEntrypointReferencePackage{
		id: identity.PackageID,
		sealed: catalogpackagev1.SealedPackage{
			Manifest: manifest, ManifestSHA256: hex.EncodeToString(digest[:]),
			Signature: base64.RawURLEncoding.EncodeToString(
				ed25519.Sign(privateKey, manifest),
			),
			KeyID: builtEntrypointCatalogPackageKeyID, OwnerPayloads: payloads,
		},
	}
}

func readBuiltEntrypointOwnerPayloads(
	t *testing.T,
	directory string,
) []catalogpackagev1.OwnerPayload {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read first-party Catalog owner artifacts: %v", err)
	}
	var result []catalogpackagev1.OwnerPayload
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "owner-") ||
			filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		document := readBuiltEntrypointCanonicalJSON(
			t, filepath.Join(directory, entry.Name()),
			catalogpackagev1.MaximumPackageSetBytes,
		)
		var group builtEntrypointOwnerGroup
		if err := json.Unmarshal(document, &group); err != nil ||
			group.Contract != "harpy.firstparty-owner-payload-group.v1" ||
			group.Owner == "" || len(group.Payloads) == 0 {
			t.Fatalf("decode first-party Catalog owner artifact %s: %v", entry.Name(), err)
		}
		for _, payload := range group.Payloads {
			if payload.Owner != group.Owner {
				t.Fatalf("owner artifact %s contains foreign payload", entry.Name())
			}
			result = append(result, payload)
		}
	}
	if len(result) == 0 {
		t.Fatalf("first-party Catalog owner payloads are missing in %s", directory)
	}
	return result
}
