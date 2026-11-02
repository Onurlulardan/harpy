package baseline

import (
	"strings"
	"testing"
)

func TestBaselineScriptsAreGooseSnapshots(t *testing.T) {
	for name, script := range map[string]string{
		"platform":         platformScript,
		"tenant":           tenantScript,
		"platform cleanup": platformCatalogRemovalScript,
		"tenant cleanup":   tenantCatalogRemovalScript,
	} {
		if strings.Contains(script, "\\restrict") || strings.Contains(script, "\\unrestrict") {
			t.Fatalf("%s baseline contains a psql meta command", name)
		}
		if !strings.Contains(script, "-- +goose Up") || !strings.Contains(script, "-- +goose StatementBegin") {
			t.Fatalf("%s baseline is not a goose statement", name)
		}
	}
	if !strings.Contains(tenantScript, tenantSchemaToken) ||
		!strings.Contains(tenantCatalogRemovalScript, tenantSchemaToken) {
		t.Fatal("tenant SQL scripts must contain the tenant schema token")
	}
	if strings.Contains(platformScript, tenantSchemaToken) ||
		strings.Contains(platformCatalogRemovalScript, tenantSchemaToken) {
		t.Fatal("platform SQL scripts must not contain the tenant schema token")
	}
}

func TestTenantSchemaName(t *testing.T) {
	if validSchema("") || validSchema("Tenant") || validSchema("1abc") || validSchema(tenantSchemaToken) {
		t.Fatal("invalid schema names were accepted")
	}
	if !validSchema("tenant_a") {
		t.Fatal("tenant_a was rejected")
	}
}

func TestCatalogRemovalMigrationsFailClosedAndRetainIAMHistory(t *testing.T) {
	for name, script := range map[string]string{
		"platform": platformCatalogRemovalScript,
		"tenant":   tenantCatalogRemovalScript,
	} {
		for _, guard := range []string{
			"requires an explicit data disposition",
			"blocked by a foreign key",
			"blocked by a view",
			"blocked by a routine",
		} {
			if !strings.Contains(script, guard) {
				t.Errorf("%s cleanup is missing %q guard", name, guard)
			}
		}
	}
	for _, statement := range []string{
		"SET state = 'revoked'",
		"SET state = 'retired'",
		"iam_resources",
		"iam_permissions",
	} {
		if !strings.Contains(tenantCatalogRemovalScript, statement) {
			t.Errorf("tenant cleanup is missing %q", statement)
		}
	}
	if strings.Contains(tenantCatalogRemovalScript, "DELETE FROM __tenant_schema__.iam_resources") ||
		strings.Contains(tenantCatalogRemovalScript, "DELETE FROM __tenant_schema__.iam_permissions") {
		t.Fatal("tenant cleanup must retain IAM history instead of hard-deleting it")
	}
}

