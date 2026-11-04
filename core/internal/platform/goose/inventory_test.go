package goose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoutineDispositionIsClosed(t *testing.T) {
	root := repositoryRoot(t)
	body, err := os.ReadFile(filepath.Join(root, "docs/tasks/05-v1-routine-disposition.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "UNRESOLVED") && !strings.Contains(text, "`UNRESOLVED`: 0") &&
		!strings.Contains(text, "- UNRESOLVED: 0") {
		t.Fatal("routine disposition still has unresolved rows")
	}
	if strings.Count(text, "UNRESOLVED") != 2 {
		t.Fatalf("unexpected UNRESOLVED mentions: %d", strings.Count(text, "UNRESOLVED"))
	}
	rows := 0
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		rows++
		if !strings.Contains(line, "PRESERVE_NON_MIGRATION") &&
			!strings.Contains(line, "REMOVE") &&
			!strings.Contains(line, "REPLACE_WITH_NEW_GOOSE") {
			t.Fatalf("row has no disposition: %s", line)
		}
	}
	if rows == 0 {
		t.Fatal("disposition table is empty")
	}
}

// yalnız negatif test: migration-streams must not exist and runtime Go must not name it.
func TestMigrationStreamCatalogIsNotRuntime(t *testing.T) {
	root := repositoryRoot(t)
	if _, err := os.Stat(filepath.Join(root, "contracts/storage/migration-streams.yaml")); err == nil {
		t.Fatal("migration stream catalog is still in the tree")
	}
	if _, err := os.Stat(filepath.Join(root, "contracts/storage/migration-streams")); err == nil {
		t.Fatal("migration stream directory is still in the tree")
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), "contracts/storage/migration-streams") {
			t.Fatalf("runtime Go references migration streams: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
