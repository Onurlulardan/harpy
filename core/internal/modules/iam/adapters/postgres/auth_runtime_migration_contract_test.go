package postgres

import (
	"strings"
	"testing"
)

func TestIAMAuthRuntimeMigrationContracts(t *testing.T) {
	tests := []struct {
		index   int
		routine string
		markers []string
	}{
		{0, "iam_login_credential_v1", []string{
			"identifier.normalized_sha256", "credential.legacy_bcrypt_hash",
			"principal.status", "binding.state", "p_source_user_id uuid",
			"p_source_identifier_id uuid",
		}},
		{1, "iam_password_use_v1", []string{
			"INTERVAL '180 days'", "algorithm = 'argon2id_v1'",
			"legacy_bcrypt_hash = NULL", "token_version =",
		}},
		{2, "iam_session_issue_v1", []string{
			"INSERT INTO iam_sessions", "INSERT INTO iam_refresh_families",
			"INSERT INTO iam_refresh_tokens", "INTERVAL '90 days'",
		}},
		{3, "iam_refresh_credential_v1", []string{
			"token.selector", "token.secret_hmac", "family.state",
			"principal.authorization_version", "p_source_token_id uuid",
			"p_source_family_id uuid", "access_mode_revision bigint",
		}},
	}
	for _, test := range tests {
		source := string(readIAMIdentityMigration(
			t, "fresh", iamAuthRuntimeMigrationNames[test.index],
		))
		for _, marker := range append([]string{
			"CREATE FUNCTION " + test.routine + "(",
			"SECURITY DEFINER", "SET search_path FROM CURRENT",
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
			"p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()",
			"REVOKE ALL ON FUNCTION " + test.routine, " FROM PUBLIC;",
		}, test.markers...) {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", test.routine, marker)
			}
		}
	}
}

func TestIAMStrictRefreshRuntimeMigrationContracts(t *testing.T) {
	for index, routine := range []string{
		"iam_refresh_decide_v2", "iam_refresh_reuse_v2", "iam_refresh_expire_v2",
		"iam_refresh_rotate_active_v2", "iam_refresh_rotate_v2",
	} {
		source := string(readIAMIdentityMigration(
			t, "fresh", iamAuthRuntimeMigrationNames[index+4],
		))
		for _, marker := range []string{
			"CREATE FUNCTION " + routine + "(", "SET search_path FROM CURRENT",
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
			"REVOKE ALL ON FUNCTION " + routine, " FROM PUBLIC;",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", routine, marker)
			}
		}
	}
	main := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[8],
	))
	for _, marker := range []string{
		"iam_refresh_decide_v2(", "locator_expected_revision",
		"iam_refresh_reuse_v2(", "iam_refresh_expire_v2(",
		"iam_refresh_rotate_active_v2(", "outcome := 'reused'",
		"outcome := 'expired'", "outcome := 'rotated'",
	} {
		if !strings.Contains(main, marker) {
			t.Fatalf("strict refresh main lacks %q", marker)
		}
	}
	acl := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[9],
	))
	if strings.Count(acl, "GRANT EXECUTE ON FUNCTION") != 5 ||
		strings.Count(acl, "REVOKE EXECUTE ON FUNCTION") != 1 ||
		strings.Contains(acl, "iam_refresh_reuse_v2(") ||
		strings.Contains(acl, "iam_refresh_expire_v2(") {
		t.Fatal("strict refresh ACL exposes internal transition helpers")
	}
}

func TestIAMStrictRefreshKeepsPlatformProjectionOutsideTenantRoutines(t *testing.T) {
	for index := 4; index <= 8; index++ {
		source := string(readIAMIdentityMigration(
			t, "fresh", iamAuthRuntimeMigrationNames[index],
		))
		if strings.Contains(source, "platform.") {
			t.Fatalf("%s calls a platform routine from tenant SQL",
				iamAuthRuntimeMigrationNames[index])
		}
	}
	decision := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[4],
	))
	for _, marker := range []string{
		"token.idle_expires_at", "access.state", "binding.state",
		"FOR UPDATE OF principal, access, session, family, token",
	} {
		if !strings.Contains(decision, marker) {
			t.Fatalf("strict refresh decision lacks %q", marker)
		}
	}
}

func TestIAMRefreshLocatorDigestRuntimeContract(t *testing.T) {
	hash := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[10],
	))
	for _, marker := range []string{
		"iam_refresh_locator_source_sha256_v1", "pg_catalog.sha256(",
		"pg_catalog.uuid_send(p_token_id)", "pg_catalog.int8send(p_revision)",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
	} {
		if !strings.Contains(hash, marker) {
			t.Fatalf("locator digest lacks %q", marker)
		}
	}
	reader := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[11],
	))
	if !strings.Contains(reader, "CREATE FUNCTION iam_refresh_credential_v2(") ||
		!strings.Contains(reader, "token.state, token.secret_hmac") {
		t.Fatal("refresh reader does not bind locator state digest")
	}
	rotation := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[12],
	))
	if strings.Count(rotation, "iam_refresh_locator_source_sha256_v1(") != 2 ||
		!strings.Contains(rotation, "FROM iam_refresh_rotate_v2(") {
		t.Fatal("refresh v3 does not derive both locator projections")
	}
	acl := string(readIAMIdentityMigration(
		t, "fresh", iamAuthRuntimeMigrationNames[13],
	))
	if strings.Count(acl, "REVOKE EXECUTE ON FUNCTION") != 2 ||
		strings.Count(acl, "GRANT EXECUTE ON FUNCTION") != 2 {
		t.Fatal("refresh v3 ACL does not retire both stale runtime routines")
	}
}
