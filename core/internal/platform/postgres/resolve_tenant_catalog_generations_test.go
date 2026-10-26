package postgres

import (
	"strings"
	"testing"
)

func TestResolveTenantCatalogGenerationsSQLRequiresExactCurrentGroup(t *testing.T) {
	for _, fragment := range []string{"group_row.scope_kind = 'tenant'",
		"group_row.group_id = 'catalog'", "catalog_member.domain = 'catalog'",
		"catalog_member.ordinal = 1", "group_row.state = 'v2'",
		"group_row.writer = 'v2'", "catalog_row.state = 'v2'",
		"catalog_row.writer = 'v2'", "extra.domain <> 'catalog'"} {
		if !strings.Contains(resolveTenantCatalogGenerationsSQL, fragment) {
			t.Fatalf("tenant Catalog generation SQL missing %q", fragment)
		}
	}
}
