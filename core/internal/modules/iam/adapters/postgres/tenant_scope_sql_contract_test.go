package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTenantAdapterSQLUsesBoundSchemaBehindPgCatalog(t *testing.T) {
	root := "."
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".sql" ||
			!strings.Contains(filepath.ToSlash(path), "/sql/tenant/") {
			return nil
		}
		document, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		source := string(document)
		if strings.Contains(source, "pg_catalog.current_schema()::name") {
			t.Errorf("%s derives the tenant scope from pg_catalog", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
