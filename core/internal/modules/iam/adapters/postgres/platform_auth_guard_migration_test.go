package postgres

import (
	"strings"
	"testing"
)

var platformAuthGuardMigrations = []string{
	"20260716012020_create_iam_tenant_authority_assertion_v1",
	"20260716012022_expand_iam_throttle_authority_epoch",
	"20260716012025_guard_iam_refresh_locators_v1",
	"20260716012030_create_iam_refresh_locator_writer_v1",
	"20260716012035_create_iam_refresh_locator_advancer_v1",
	"20260716012040_guard_iam_auth_throttle_v1",
	"20260716012042_create_iam_auth_throttle_policy_v1",
	"20260716012045_create_iam_tenant_throttle_writer_v1",
	"20260716012050_create_iam_environment_throttle_writer_v1",
	"20260716012055_grant_iam_platform_guard_routines_v1",
	"20260716012056_create_iam_tenant_target_resolver_v1",
	"20260716012057_create_iam_login_target_resolver_v1",
	"20260716012058_create_iam_refresh_target_resolver_v1",
	"20260716012059_grant_iam_target_resolvers_v1",
	"20260716013340_create_iam_login_target_resolver_v2",
	"20260716013341_create_iam_refresh_target_resolver_v2",
	"20260716013342_grant_iam_target_resolvers_v2",
}

func assertIAMTargetResolverMigrationContract(t *testing.T) {
	t.Helper()
	for index, markers := range map[int][]string{
		10: {"iam_resolve_tenant_target_v1(p_tenant_id uuid)", "FOR SHARE"},
		11: {"iam_resolve_login_target_v1(", "p_tenant_id uuid", "pg_advisory_xact_lock", "INTO STRICT"},
		12: {"iam_resolve_refresh_target_v1(p_selector bytea)", "retain_until", "FOR SHARE"},
	} {
		source := guardMigrationSource(t, platformAuthGuardMigrations[index])
		assertPlatformAuthMarkers(t, source, append(markers,
			"SECURITY DEFINER", "SET search_path = pg_catalog", "REVOKE ALL ON FUNCTION"))
	}
	grants := guardMigrationSource(t, platformAuthGuardMigrations[13])
	if strings.Count(grants, "GRANT EXECUTE ON FUNCTION") != 3 ||
		strings.Contains(grants, "GRANT SELECT") {
		t.Fatal("target resolver ACL must expose exactly three routines")
	}
	loginV2 := guardMigrationSource(t, platformAuthGuardMigrations[14])
	assertPlatformAuthMarkers(t, loginV2, []string{
		"iam_resolve_login_target_v2(", "source_user_id uuid",
		"source_identifier_id uuid", "source_identifier_kind text", "FOR SHARE",
	})
	refreshV2 := guardMigrationSource(t, platformAuthGuardMigrations[15])
	assertPlatformAuthMarkers(t, refreshV2, []string{
		"iam_resolve_refresh_target_v2", "source_token_id uuid",
		"source_family_id uuid", "source_selector bytea", "FOR UPDATE",
	})
	seal := guardMigrationSource(t, platformAuthGuardMigrations[16])
	if strings.Count(seal, "GRANT EXECUTE ON FUNCTION") != 2 ||
		strings.Count(seal, "REVOKE EXECUTE ON FUNCTION") != 2 {
		t.Fatal("v2 target resolver ACL does not retire both weak resolvers")
	}
}

func guardMigrationSource(t *testing.T, name string) string {
	t.Helper()
	return string(readIAMPlatformAuthMigration(t, "fresh", name))
}
