package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinalizerUsesVersionedTenantRequirementAuthority(t *testing.T) {
	finalizers := cutoverMigrationDocuments(t,
		"20260714014000_create_authority_finalize_operation_v1",
	)
	for path, document := range finalizers {
		assertNoDirectTenantRequirementRead(t, path, document)
		for _, fragment := range []string{
			"cutover_authority.validate_finalization_envelope_v1(",
			"platform.tenant_migration_requirement_catalogs AS catalog",
			"tenant_requirement_catalog_document",
		} {
			if !strings.Contains(document, fragment) {
				t.Fatalf("protected finalizer %s lacks %q", path, fragment)
			}
		}
	}

	dispatchers := cutoverMigrationDocuments(t,
		"20260714013759_create_terminal_evidence_assertion",
	)
	for path, document := range dispatchers {
		assertNoDirectTenantRequirementRead(t, path, document)
		if !strings.Contains(document,
			"cutover_authority.assert_tenant_completion_v1(") {
			t.Fatalf("terminal dispatcher %s lacks protected tenant assertion", path)
		}
	}

	tenantAssertions := cutoverMigrationDocuments(t,
		"20260714013758_create_tenant_completion_assertion",
	)
	for path, document := range tenantAssertions {
		for _, fragment := range []string{
			"tenant-migration-requirement-catalog.v2",
			"platform.tenant_migration_requirement_targets",
			"platform.tenant_migration_checkpoints",
			"'tenant_id',live.tenant_id,'schema_name',live.schema_name",
		} {
			if !strings.Contains(document, fragment) {
				t.Fatalf("tenant authority assertion %s lacks %q", path, fragment)
			}
		}
	}
}

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
