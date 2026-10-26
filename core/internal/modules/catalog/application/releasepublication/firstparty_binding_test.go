package releasepublication

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

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

type firstPartyOwnerGroup struct {
	Contract string                          `json:"contract"`
	Owner    string                          `json:"owner"`
	Payloads []catalogpackagev1.OwnerPayload `json:"payloads"`
}

func TestPackageVersionsBindEveryFirstPartyOwnerArtifact(t *testing.T) {
	modules := []string{
		"hr", "procurement", "sales-crm", "accounting", "inventory",
		"quality", "production", "maintenance", "project",
	}
	root := firstPartyRepositoryRoot(t)
	set := catalogpackagev1.ReleaseSet{
		ContractVersion: catalogpackagev1.ContractVersion,
		ReleaseID:       "018f0000-0006-7000-8000-000000000001",
	}
	for _, module := range modules {
		set.Packages = append(
			set.Packages, loadFirstPartySealedPackage(t, root, module),
		)
	}
	versions, err := PackageVersionsFromReleaseSet(set)
	if err != nil {
		t.Fatalf("bind first-party package set: %v", err)
	}
	totalPayloads := 0
	for _, version := range versions {
		totalPayloads += len(version.OwnerPayloads)
	}
	if len(versions) != 9 || totalPayloads != 466 {
		t.Fatalf(
			"bound first-party coverage = %d packages/%d payloads, want 9/466",
			len(versions), totalPayloads,
		)
	}
}

func loadFirstPartySealedPackage(
	t *testing.T, root, module string,
) catalogpackagev1.SealedPackage {
	t.Helper()
	directory := filepath.Join(
		root, "tools", "internal", "firstpartycatalog", "testdata",
		module, "artifacts",
	)
	manifest, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatalf("read %s manifest: %v", module, err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read %s artifacts: %v", module, err)
	}
	var payloads []catalogpackagev1.OwnerPayload
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "owner-") ||
			!strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatalf("read %s %s: %v", module, entry.Name(), readErr)
		}
		var group firstPartyOwnerGroup
		if json.Unmarshal(document, &group) != nil ||
			group.Contract != "harpy.firstparty-owner-payload-group.v1" ||
			group.Owner == "" || len(group.Payloads) == 0 {
			t.Fatalf("%s %s owner group is invalid", module, entry.Name())
		}
		for _, payload := range group.Payloads {
			if payload.Owner != group.Owner {
				t.Fatalf("%s %s payload owner differs", module, entry.Name())
			}
			payloads = append(payloads, payload)
		}
	}
	sort.Slice(payloads, func(left, right int) bool {
		return payloads[left].ComponentID < payloads[right].ComponentID
	})
	digest := sha256.Sum256(manifest)
	return catalogpackagev1.SealedPackage{
		Manifest:       manifest,
		ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature: base64.RawURLEncoding.EncodeToString(
			make([]byte, ed25519.SignatureSize),
		),
		KeyID:         "catalog-test-2026-07",
		OwnerPayloads: payloads,
	}
}

func firstPartyRepositoryRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("backend-v2 repository root was not found")
		}
		current = parent
	}
}
