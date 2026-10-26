package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestCreateTenantRequirementCatalogRejectsCountMismatchBeforeWrite(t *testing.T) {
	catalog := TenantRequirementCatalog{
		RequirementCount: 2,
		Requirements:     []TenantMigrationRequirement{{Ordinal: 1}},
	}

	err := CreateTenantRequirementCatalog(context.Background(), nil, catalog)
	if err == nil || !strings.Contains(err.Error(), "count differs") {
		t.Fatalf("count mismatch error = %v", err)
	}
}
