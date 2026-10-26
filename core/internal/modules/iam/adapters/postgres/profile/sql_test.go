package profile

import (
	"strings"
	"testing"
)

func TestCommandSQLUsesProductionTenantScopeAndVersionedRoutines(t *testing.T) {
	documents := map[string]struct{ sql, version string }{
		"publish": {publishSQL, "_v1("}, "backfill": {backfillSQL, "_v1("},
		"retire": {retireSQL, "_v1("}, "get_self": {getSelfSQL, "_v1("},
		"update_self": {updateSelfSQL, "_v2("},
	}
	for name, document := range documents {
		if !strings.Contains(document.sql, "(pg_catalog.current_schemas(false))[2]::name") ||
			strings.Contains(document.sql, "current_schema()") ||
			!strings.Contains(document.sql, document.version) {
			t.Fatalf("%s SQL has an unsafe tenant routine call", name)
		}
	}
}
