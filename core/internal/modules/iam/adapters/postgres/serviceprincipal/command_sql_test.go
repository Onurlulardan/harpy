package serviceprincipal

import (
	"strings"
	"testing"
)

func TestCommandSQLUsesProductionTenantScopeAndVersionedRoutines(t *testing.T) {
	documents := map[string]string{
		"create": createSQL, "disable": disableSQL, "enable": enableSQL,
		"retire": retireSQL, "grant": grantSQL, "revoke": revokeSQL,
	}
	for name, document := range documents {
		if !strings.Contains(document, "(pg_catalog.current_schemas(false))[2]::name") ||
			strings.Contains(document, "current_schema()") ||
			!strings.Contains(document, "_v1(") {
			t.Fatalf("%s SQL has an unsafe tenant routine call", name)
		}
	}
}
