package cutover

import (
	"path"
	"strings"
	"testing"
)

var tenantSingletonAuthorityVersions = []string{
	"20260728103000",
	"20260728103001",
	"20260728103002",
}

func TestTenantSingletonAuthorityMigrationIsPairedAndForwardOnly(t *testing.T) {
	for _, version := range tenantSingletonAuthorityVersions {
		source := tenantSingletonAuthorityMigrationPair(t, version)["fresh"]
		if strings.Count(source, "-- +goose Up") != 1 ||
			strings.Contains(source, "-- +goose Down") ||
			strings.Contains(source, "CREATE OR REPLACE") {
			t.Fatalf("%s migration is not immutable and forward-only", version)
		}
	}
}

func TestTenantSingletonAuthorityMigrationCoversEveryRuntimeGroup(t *testing.T) {
	backfill := tenantSingletonAuthorityMigrationPair(
		t, tenantSingletonAuthorityVersions[0],
	)["fresh"]
	initializer := tenantSingletonAuthorityMigrationPair(
		t, tenantSingletonAuthorityVersions[1],
	)["fresh"]
	compatibility := tenantSingletonAuthorityMigrationPair(
		t, tenantSingletonAuthorityVersions[2],
	)["fresh"]
	singletons := []string{
		"analytics", "catalog", "documents", "notifications", "work", "workflow",
	}
	for _, singleton := range singletons {
		if !strings.Contains(backfill, "('"+singleton+"')") ||
			!strings.Contains(initializer, "'"+singleton+"'") {
			t.Fatalf("tenant singleton authority migrations lack %s", singleton)
		}
	}
	for _, marker := range []string{
		"CREATE FUNCTION platform.cutover_initialize_tenant_v2(",
		"<> 8", "<> 10",
		"tenant cutover initialization v2 domain conflict",
		"tenant cutover initialization v2 authority conflict",
	} {
		if !strings.Contains(initializer, marker) {
			t.Fatalf("tenant singleton initializer lacks %q", marker)
		}
	}
	if !strings.Contains(compatibility,
		"CREATE TRIGGER cutover_sync_singleton_authority_insert_v1") ||
		!strings.Contains(compatibility,
			"'analytics', 'catalog', 'documents', 'work', 'workflow'") ||
		strings.Contains(compatibility, "'notifications'") {
		t.Fatal("legacy runtime compatibility trigger has the wrong singleton scope")
	}
}

func tenantSingletonAuthorityMigrationPair(
	t *testing.T, version string,
) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pair := map[string]string{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if strings.HasPrefix(name, version+"_") && strings.HasSuffix(name, ".fresh.sql") {
			pair["fresh"] = string(document.Content())
		}
	}
	if pair["fresh"] == "" {
		t.Fatalf("fresh migration %s did not load", version)
	}
	return pair
}
