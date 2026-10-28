package migrator

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func TestFilesystemBundleMatchesRuntimeFreshRegistry(t *testing.T) {
	documents, err := deployMigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	registry, err := migration.LoadStartupRegistry(documents)
	if err != nil {
		t.Fatal(err)
	}
	digests := registry.Digests()
	if len(digests) != 1 {
		t.Fatalf("fresh digests = %d", len(digests))
	}
	artifact, err := registry.MigrationArtifact(digests[0])
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := migrationbundle.CanonicalFresh(moduleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(artifact.CanonicalJSON(), canonical) {
		t.Fatalf("fresh producers differ: runtime %d bytes filesystem %d bytes",
			len(artifact.CanonicalJSON()), len(canonical))
	}
	again, err := migrationbundle.CanonicalFresh(moduleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical, again) {
		t.Fatal("filesystem fresh manifest is not deterministic")
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root is unavailable")
		}
		dir = parent
	}
}
