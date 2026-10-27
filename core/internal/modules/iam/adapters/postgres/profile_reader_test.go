package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestProfileReaderSQLUsesOnlyTenantScopedReadRoutines(t *testing.T) {
	tests := []struct {
		document string
		routine  string
	}{
		{profileHistoricalSchemaSQL, "iam_profile_historical_read_v1"},
		{profileActiveCatalogSQL, "iam_profile_active_catalog_read_v1"},
		{profileHistoricalDocumentSQL, "iam_profile_document_read_v1"},
	}
	for _, test := range tests {
		if !strings.Contains(test.document, test.routine) ||
			!strings.Contains(test.document,
				"(pg_catalog.current_schemas(false))[2]::name") {
			t.Fatalf("profile SQL did not bind routine %s to tenant scope", test.routine)
		}
		for _, direct := range []string{
			"FROM iam_users ", "FROM iam_roles ", "FROM iam_organizations ",
			"FROM iam_profile_schema_versions ",
			"FROM iam_profile_field_revisions ", "current_schema()",
		} {
			if strings.Contains(test.document, direct) {
				t.Fatalf("profile SQL directly opens an IAM table: %s", direct)
			}
		}
	}
}

func TestProfileReaderRequiresBoundQueryAndScopeCheck(t *testing.T) {
	if (&profileReader{}).usable(context.Background()) {
		t.Fatal("profile reader became usable without its TenantScope binder")
	}
}
