package goose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoutineDispositionIsClosed(t *testing.T) {
	root := repositoryRoot(t)
	catalog := filepath.Join(root, "docs", "tasks", "05-v1-routine-disposition.md")
	if _, err := os.Stat(catalog); err == nil {
		t.Fatal("retired routine disposition catalog is present")
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".sql":
		default:
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(body)
		if strings.Contains(text, "05-v1-routine-disposition") || strings.Contains(text, "UNRESOLVED") {
			t.Fatalf("open routine disposition remains in %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
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
