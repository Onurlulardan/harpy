package postgres

import "testing"

func TestTenantSearchPathQuotesOnlyCanonicalRegistryNames(t *testing.T) {
	value, err := tenantSearchPath("tenant_00112233445566778899aabbccddeeff")
	if err != nil || value != `pg_catalog,"tenant_00112233445566778899aabbccddeeff"` {
		t.Fatalf("unexpected tenant search path %q: %v", value, err)
	}
	for _, invalid := range []string{"", "public,pg_temp", "Tenant", "has-dash", "a\x00b"} {
		if _, err := tenantSearchPath(invalid); err == nil {
			t.Fatalf("invalid schema name %q accepted", invalid)
		}
	}
}
