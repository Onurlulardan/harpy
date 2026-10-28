package migrationbundle

import (
	"archive/zip"
	"path/filepath"
	"testing"
)

func TestReleaseNamesMatchIndex(t *testing.T) {
	root := moduleRoot(t)
	if err := ReleaseNamesMatchIndex(root); err != nil {
		t.Fatal(err)
	}
}

func TestPackReleaseWritesNamedMembers(t *testing.T) {
	root := moduleRoot(t)
	output := filepath.Join(t.TempDir(), "migration-release.zip")
	if err := PackRelease(root, output); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	digests, err := indexedDigests(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.File) != len(digests) {
		t.Fatalf("zip members = %d, digests = %d", len(reader.File), len(digests))
	}
}
