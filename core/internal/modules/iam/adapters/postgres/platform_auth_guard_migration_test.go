package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
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

func TestIAMPlatformAuthGuardMigrationsStayPairedAndBounded(t *testing.T) {
	for _, name := range platformAuthGuardMigrations {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMPlatformAuthMigration(t, "fresh", name)
			adopted := readIAMPlatformAuthMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted guard migration differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("platform auth guard migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("inspect platform auth guard migration: %v", err)
			}
			source := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"-- +GOOSE DOWN", "CREATE OR REPLACE",
				"GRANT INSERT", "GRANT UPDATE", "GRANT DELETE",
				"GRANT SELECT", "GRANT ALL ON TABLE",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("guard migration contains %q", forbidden)
				}
			}
		})
	}
}

func TestIAMTenantAuthorityAssertionContract(t *testing.T) {
	source := guardMigrationSource(t, platformAuthGuardMigrations[0])
	markers := []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"FROM platform.cutover_groups AS authority",
		"authority.group_id OPERATOR(pg_catalog.=) 'iam_system_resources'",
		"FROM platform.domain_cutovers AS authority",
		"authority.domain OPERATOR(pg_catalog.=) 'iam'",
		"authority.domain OPERATOR(pg_catalog.=) 'system_resource_layout'",
		"v_group_state IS DISTINCT FROM 'v2'",
		"v_group_writer IS DISTINCT FROM 'v2'",
		"v_domain_generation IS DISTINCT FROM v_group_generation",
		"member.ordinal OPERATOR(pg_catalog.=) 1",
		"member.ordinal OPERATOR(pg_catalog.=) 2",
		"FOR SHARE", "SECURITY DEFINER", "SET search_path = pg_catalog",
	}
	assertPlatformAuthMarkers(t, source, markers)
	if strings.Count(source, "FOR SHARE") != 3 {
		t.Fatal("authority assertion must lock group and both exact domains")
	}
}

func TestIAMPlatformGuardAndWriterContracts(t *testing.T) {
	epoch := guardMigrationSource(t, platformAuthGuardMigrations[1])
	assertPlatformAuthMarkers(t, epoch, []string{
		"ADD COLUMN authority_epoch bigint NOT NULL",
		"purpose, dimension,\n        key_id, bucket_hmac, authority_epoch",
		"authority_epoch = authority_group_generation",
		"authority_epoch = cutover_generation",
		"bucket_scope_kind = 'environment' AND authority_epoch = 0",
	})
	locatorGuard := guardMigrationSource(t, platformAuthGuardMigrations[2])
	assertPlatformAuthMarkers(t, locatorGuard, []string{
		"BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_refresh_locators",
		"ENABLE ALWAYS TRIGGER iam_refresh_locators_guard",
		"IAM refresh locator deletion is forbidden",
		"NEW.source_revision OPERATOR(pg_catalog.<=) OLD.source_revision",
		"NEW.retain_until OPERATOR(pg_catalog.<) OLD.retain_until",
	})
	if strings.Contains(locatorGuard, "SECURITY DEFINER") {
		t.Fatal("locator trigger must remain security invoker")
	}
	throttleGuard := guardMigrationSource(t, platformAuthGuardMigrations[5])
	assertPlatformAuthMarkers(t, throttleGuard, []string{
		"BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_auth_throttle_buckets",
		"ENABLE ALWAYS TRIGGER iam_auth_throttle_buckets_guard",
		"FROM platform.environment_identity AS environment",
		"NEW.revision IS DISTINCT FROM",
		"NEW.attempt_count IS DISTINCT FROM",
		"IAM throttle active block cannot shrink",
		"NEW.window_start OPERATOR(pg_catalog.<) OLD.window_end",
	})
	if strings.Contains(throttleGuard, "SECURITY DEFINER") {
		t.Fatal("throttle trigger must remain security invoker")
	}
	for _, index := range []int{3, 4, 7, 8} {
		source := guardMigrationSource(t, platformAuthGuardMigrations[index])
		assertPlatformAuthMarkers(t, source, []string{
			"SECURITY DEFINER", "SET search_path = pg_catalog",
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"REVOKE ALL ON FUNCTION",
		})
	}
	environment := guardMigrationSource(t, platformAuthGuardMigrations[8])
	if strings.Contains(environment, "p_environment_id") ||
		!strings.Contains(environment, "FROM platform.environment_identity") {
		t.Fatal("environment writer must derive and lock exact environment identity")
	}
	grants := guardMigrationSource(t, platformAuthGuardMigrations[9])
	if strings.Count(grants, "GRANT EXECUTE ON FUNCTION") != 4 ||
		strings.Contains(grants, "GRANT INSERT") || strings.Contains(grants, "GRANT UPDATE") {
		t.Fatal("platform guard ACL must expose only four narrow writer routines")
	}
	policy := guardMigrationSource(t, platformAuthGuardMigrations[6])
	assertPlatformAuthMarkers(t, policy, []string{
		"iam_auth_throttle_policy_v1", "IMMUTABLE", "PARALLEL SAFE",
		"p_policy_version IS NOT DISTINCT FROM 1", "REVOKE ALL ON FUNCTION",
	})
	tenantThrottle := guardMigrationSource(t, platformAuthGuardMigrations[7])
	assertPlatformAuthMarkers(t, tenantThrottle, []string{
		"ORDER BY bucket.authority_epoch DESC LIMIT 1 FOR SHARE",
		"IAM tenant throttle prior policy is newer",
		"v_now OPERATOR(pg_catalog.<=) v_row.last_attempt_at",
	})
	assertIAMTargetResolverMigrationContract(t)
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
