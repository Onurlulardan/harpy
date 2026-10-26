package postgres

import (
	"strings"
	"testing"
)

func TestTenantSlugLookupAcceptsExactLegacyAndCurrentSlugs(t *testing.T) {
	for _, slug := range []string{"acme", "legacy_tenant", "musteri-01", "musteri"} {
		if !validTenantSlugLookup(slug) {
			t.Fatalf("valid tenant slug rejected: %q", slug)
		}
	}
}

func TestTenantSlugLookupRejectsAmbiguousOrUnboundedInput(t *testing.T) {
	for _, slug := range []string{
		"", " acme", "acme ", "acme\nother", strings.Repeat("a", 101),
		string([]byte{0xff}),
	} {
		if validTenantSlugLookup(slug) {
			t.Fatalf("invalid tenant slug accepted: %q", slug)
		}
	}
}

func TestTenantSlugLookupUsesNarrowRuntimeResolver(t *testing.T) {
	if !strings.Contains(resolveTenantSlugSQL, "tenancy_runtime_resolve_slug_v1") ||
		strings.Contains(resolveTenantSlugSQL, "FROM platform.tenants") {
		t.Fatal("tenant slug lookup bypasses the narrow runtime resolver")
	}
}
