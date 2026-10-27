package postgres

import "testing"

func TestTenantSearchPathQuotesOnlyCanonicalRegistryNames(t *testing.T) {
	value, err := tenantSearchPath("t_abcdefghijklmnopqrstuvwxyz")
	if err != nil || value != `pg_catalog,"t_abcdefghijklmnopqrstuvwxyz"` {
		t.Fatalf("unexpected tenant search path %q: %v", value, err)
	}
	for _, invalid := range []string{"", "public,pg_temp", "Tenant", "has-dash", "a\x00b"} {
		if _, err := tenantSearchPath(invalid); err == nil {
			t.Fatalf("invalid schema name %q accepted", invalid)
		}
	}
}
