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
