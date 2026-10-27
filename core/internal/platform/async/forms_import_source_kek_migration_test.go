package async

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormsImportSourceKEKMigrationPairs(t *testing.T) {
	for _, name := range []string{
		"20260718020110_create_forms_import_source_kek_inventory_v1",
		"20260718020111_grant_forms_import_source_kek_inventory_v1",
	} {
		fresh, err := os.ReadFile(filepath.Join("migrations", "fresh", name+".fresh.sql"))
		if err != nil {
			t.Fatal(err)
		}
		adopted, err := os.ReadFile(filepath.Join(
			"migrations", "v1_adopted", name+".v1_adopted.sql",
		))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("migration pair differs: %s", name)
		}
	}
	document, err := os.ReadFile(filepath.Join("migrations", "fresh",
		"20260718020110_create_forms_import_source_kek_inventory_v1.fresh.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{
		"platform_async_forms_import_source_kek_tenant_v1",
		"authority.domain='forms'", "tenant.deleted_at IS NULL", "LIMIT 1",
	} {
		if !strings.Contains(string(document), token) {
			t.Fatalf("KEK inventory migration misses %q", token)
		}
	}
}
