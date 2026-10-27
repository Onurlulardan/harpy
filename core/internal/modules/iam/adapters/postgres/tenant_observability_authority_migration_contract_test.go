package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const iamTenantObservabilityAuthorityMigration = "20260802234200_publish_iam_tenant_observability_authority_v1"

func TestIAMTenantObservabilityAuthorityMigrationContract(t *testing.T) {
	fresh := readIAMIdentityMigration(
		t, "fresh", iamTenantObservabilityAuthorityMigration,
	)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM tenant observability authority migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM tenant observability authority migration: %v", err)
	}
	assertIAMRBACMigrationSource(t, fresh)
	assertIAMTenantObservabilityCatalogSource(t, string(fresh))
	assertIAMTenantObservabilityGrantSource(t, string(fresh))
}

func assertIAMTenantObservabilityCatalogSource(t *testing.T, source string) {
	t.Helper()
	for _, marker := range []string{
		"'changelog'::text, 'audit'::text, 'audit_changelog'::text",
		"'analytics.usage', 'usage', 'usage_analytics'",
		"'read'::text, 'Read'::text",
		"MERGE INTO iam_permissions AS target",
		"target.match_kind='exact'",
		"target.is_protected IS DISTINCT FROM true",
		"iam_tenant_observability_authority_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM tenant observability authority lacks %q", marker)
		}
	}
}

func assertIAMTenantObservabilityGrantSource(t *testing.T, source string) {
	t.Helper()
	start := strings.Index(source, "FROM (VALUES\n        ('changelog'::text")
	if start < 0 {
		t.Fatal("IAM tenant observability grant map is absent")
	}
	end := strings.Index(source[start:], ") AS grant_map(resource_key,role_name)")
	if end < 0 {
		t.Fatal("IAM tenant observability grant map is unterminated")
	}
	grantMap := source[start : start+end]
	if strings.Count(grantMap, "('") != 8 {
		t.Fatalf("IAM tenant observability grant tuple count = %d, want 8",
			strings.Count(grantMap, "('"))
	}
	for _, tuple := range []string{
		"('changelog'::text,'Tenant Yöneticisi'::text)",
		"('changelog','Genel Müdür')", "('changelog','Temel Kullanıcı')",
		"('changelog','Yönetici')", "('changelog','Müdür')",
		"('changelog','Kullanıcı')",
		"('analytics.usage','Tenant Yöneticisi')",
		"('analytics.usage','Yönetici')",
	} {
		if !strings.Contains(grantMap, tuple) {
			t.Fatalf("IAM tenant observability grant map lacks %q", tuple)
		}
	}
	for _, forbidden := range []string{"Tenant Admin", "General Manager", "Basic User"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM tenant observability authority guesses role %q", forbidden)
		}
	}
	for _, marker := range []string{
		"MERGE INTO iam_role_permission_bindings AS target",
		"WHERE role.status='active' AND NOT role.is_protected",
		"UPDATE iam_authorization_clock",
		"SELECT iam_rbac_bump_all_users_v1(clock_value) FROM bumped",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM tenant observability grant source lacks %q", marker)
		}
	}
}
