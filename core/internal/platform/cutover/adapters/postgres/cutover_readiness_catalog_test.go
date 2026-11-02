package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cutoverMigrationDocuments(t *testing.T, base string) map[string]string {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", "fresh", base+".fresh.sql")
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(document) == 0 {
		t.Fatalf("fresh migration %s did not load", base)
	}
	return map[string]string{path: string(document)}
}

func assertNoDirectTenantRequirementRead(
	t *testing.T,
	path string,
	document string,
) {
	t.Helper()
	for _, forbidden := range []string{
		"platform.tenant_migration_checkpoints",
		"platform.tenants",
		"core/internal/modules/tenancy/migrations/tenant/",
	} {
		if strings.Contains(document, forbidden) {
			t.Fatalf("finalizer %s contains direct tenant authority %q", path, forbidden)
		}
	}
}
