package postgres

import (
	"strings"
	"testing"
)

func TestGlobalBillingSQLHardCodesExactScopeAndLockMode(t *testing.T) {
	locks := []string{
		lockGlobalBillingGroupSQL,
		lockGlobalBillingMembersSQL,
		lockGlobalBillingDomainSQL,
	}
	for _, document := range locks {
		if !strings.Contains(document, "('global', $1, 'billing')") ||
			strings.Contains(document, "$2") {
			t.Fatalf("global command SQL is not exact: %q", document)
		}
	}
	for _, document := range []string{
		verifyGlobalBillingGroupSQL,
		verifyGlobalBillingMembersSQL,
		verifyGlobalBillingDomainSQL,
	} {
		if !strings.Contains(document, "scope_kind = 'global'") ||
			strings.Contains(document, "FOR SHARE") || strings.Contains(document, "$2") {
			t.Fatalf("global snapshot SQL is not exact read-only SQL: %q", document)
		}
	}
}

func TestGlobalIdentitySQLPinsPgCatalogAndRegisteredSingleton(t *testing.T) {
	for _, fragment := range []string{
		"set_config('search_path', 'pg_catalog', true)",
		"pg_catalog.pg_settings",
	} {
		if !strings.Contains(setGlobalSearchPathSQL, fragment) {
			t.Fatalf("global search_path SQL misses %q", fragment)
		}
	}
	if !strings.Contains(resolveEnvironmentIDSQL, "platform.environment_identity") ||
		!strings.Contains(resolveEnvironmentIDSQL, "WHERE singleton") ||
		!strings.Contains(verifyGlobalEnvironmentIDSQL, "environment_id = $1") {
		t.Fatal("global environment identity SQL is not registry-bound")
	}
}
