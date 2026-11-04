package goose

import (
	"io/fs"
	"strings"
	"testing"
)

func TestSourceOmitsLegacyCorpus(t *testing.T) {
	for _, root := range []struct {
		files fs.FS
		dir   string
	}{{platformRoot, "sql/platform"}, {tenantRoot, "sql/tenant"}} {
		source, err := fs.Sub(root.files, root.dir)
		if err != nil {
			t.Fatal(err)
		}
		err = fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() {
				return walkErr
			}
			if strings.Contains(path, "fresh") || strings.Contains(path, "20261028") ||
				strings.Contains(path, "20261102") || strings.Contains(path, "20261103") {
				t.Fatalf("legacy migration name %s", path)
			}
			body, err := fs.ReadFile(source, path)
			if err != nil {
				return err
			}
			text := string(body)
			// PRESERVE_NON_MIGRATION: Goose SQL must not create the native history table itself.
			if strings.Contains(text, "goose_db_version") || strings.Contains(text, "CREATE FUNCTION") {
				t.Fatalf("legacy body in %s", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
