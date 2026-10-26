//go:build integration

package migrator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

func writeBuiltEntrypointCatalogPackageSet(
	t *testing.T,
	directory string,
	name string,
	document []byte,
) builtEntrypointArtifact {
	t.Helper()
	if len(document) == 0 || len(document) > catalogpackagev1.MaximumPackageSetBytes {
		t.Fatalf("built entrypoint Catalog package set size = %d", len(document))
	}
	digest := sha256.Sum256(document)
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatalf("write built entrypoint Catalog package set: %v", err)
	}
	return builtEntrypointArtifact{
		path: path, digest: hex.EncodeToString(digest[:]), document: document,
	}
}

func TestBuiltEntrypointCatalogPackageSetUsesAllFirstPartyArtifacts(t *testing.T) {
	stage := builtEntrypointStage{directory: builtEntrypointRepositoryRoot(t)}
	artifact, keyring := prepareBuiltEntrypointCatalogPackageSet(
		t, t.TempDir(), stage,
	)
	set, err := catalogpackagev1.DecodeCanonical(artifact.document)
	if err != nil || len(set.Packages) != 9 {
		t.Fatalf("built entrypoint Catalog package set = %d packages, err=%v",
			len(set.Packages), err)
	}
	if !strings.Contains(keyring, builtEntrypointCatalogPackageKeyID) {
		t.Fatalf("built entrypoint Catalog package keyring omits signer")
	}
	for _, sealed := range set.Packages {
		if sealed.KeyID != builtEntrypointCatalogPackageKeyID {
			t.Fatalf("built entrypoint Catalog package uses foreign signer")
		}
	}
}

func readBuiltEntrypointCanonicalJSON(
	t *testing.T,
	path string,
	limit int,
) []byte {
	t.Helper()
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read canonical first-party artifact %s: %v", path, err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, limit)
	if err != nil || !bytes.Equal(canonical, document) {
		t.Fatalf("first-party artifact is not canonical %s: %v", path, err)
	}
	return document
}
