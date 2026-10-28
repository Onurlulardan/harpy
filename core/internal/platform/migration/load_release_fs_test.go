package migration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseIndexMatchesManifests(t *testing.T) {
	built, err := BuildReleaseIndex(os.DirFS("release"), ".")
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("HARPY_WRITE_RELEASE_INDEX") == "1" {
		if err = os.WriteFile("release-index.json", built, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	embedded, err := os.ReadFile("release-index.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(built, embedded) {
		t.Fatal("release index is stale; rerun with HARPY_WRITE_RELEASE_INDEX=1")
	}
}

func TestReleaseArchiveAuthorizesIndexedDigest(t *testing.T) {
	registry, err := LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	digests := registry.Digests()
	if len(digests) < 3 {
		t.Fatalf("digest count %d", len(digests))
	}
	for _, index := range []int{0, len(digests) / 2, len(digests) - 1} {
		artifact, artifactErr := registry.MigrationArtifact(digests[index])
		if artifactErr != nil {
			t.Fatal(artifactErr)
		}
		if err = registry.Authorize(artifact); err != nil {
			t.Fatalf("authorize %s: %v", digests[index], err)
		}
	}
	missing := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err = registry.AuthorizeDigest(missing); err == nil {
		t.Fatal("unknown digest was authorized")
	}
	if _, err = registry.MigrationArtifact(missing); err == nil {
		t.Fatal("unknown digest returned a manifest")
	}
}

func TestMutatedManifestFailsAuthorize(t *testing.T) {
	entries, err := os.ReadDir("release")
	if err != nil {
		t.Fatal(err)
	}
	var name string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), releaseManifestSuffix) {
			name = entry.Name()
			break
		}
	}
	body, err := os.ReadFile(filepath.Join("release", name))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err = os.WriteFile(filepath.Join(directory, name), body, 0o644); err != nil {
		t.Fatal(err)
	}
	index, err := BuildReleaseIndex(os.DirFS(directory), ".")
	if err != nil {
		t.Fatal(err)
	}
	body[0] ^= 0xff
	if err = os.WriteFile(filepath.Join(directory, name), body, 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := loadIndexedRelease(index, directoryRelease{root: directory})
	if err != nil {
		t.Fatal(err)
	}
	digest := strings.TrimSuffix(name, releaseManifestSuffix)
	if _, err = registry.MigrationArtifact(digest); err == nil {
		t.Fatal("mutated manifest was authorized")
	}
}
