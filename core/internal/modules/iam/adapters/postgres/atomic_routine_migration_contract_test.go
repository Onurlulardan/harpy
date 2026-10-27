package postgres

import (
	"strings"
	"testing"
)

func TestIAMAtomicRoutineMigrationContracts(t *testing.T) {
	tests := []struct {
		index   int
		routine string
		markers []string
	}{
		{0, "iam_refresh_rotate_v1", []string{
			"FOR UPDATE OF principal, session, family, token",
			"RETURN 'rotated'", "RETURN 'reused'", "reuse_detected",
			"token.auth_time", "token.step_up_purpose_sha256",
		}},
		{1, "iam_one_time_consume_v1", []string{
			"FOR UPDATE", "password_reset_v1", "email_verification_v1",
			"RETURN 'consumed'", "RETURN 'expired'",
		}},
		{2, "iam_invitation_accept_v1", []string{
			"p_expected_invitation_revision", "p_expected_token_revision",
			"accepted_identifier_id", "RETURN 'accepted'",
		}},
		{3, "iam_challenge_resolve_v1", []string{
			"p_expected_attempt_count", "RETURN 'pending'",
			"RETURN 'failed'", "RETURN 'consumed'",
		}},
		{4, "iam_recovery_consume_v1", []string{
			"FOR UPDATE OF principal, authenticator, code",
			"renewal_required", "token_version",
		}},
		{5, "iam_portal_setup_complete_v1", []string{
			"p_expected_setup_revision", "portal_setup_revision",
			"RETURN 'completed'", "RETURN 'expired'",
		}},
	}
	for _, test := range tests {
		source := string(readIAMIdentityMigration(
			t, "fresh", iamAtomicRoutineMigrationNames[test.index],
		))
		for _, marker := range append([]string{
			"CREATE FUNCTION " + test.routine + "(",
			"SECURITY DEFINER", "SET search_path FROM CURRENT",
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
			"p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()",
			"OPERATOR(pg_catalog.", "REVOKE ALL ON FUNCTION " + test.routine,
			" FROM PUBLIC;",
		}, test.markers...) {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", test.routine, marker)
			}
		}
	}
	grants := string(readIAMIdentityMigration(
		t, "fresh", iamAtomicRoutineMigrationNames[6],
	))
	if strings.Count(grants, "GRANT EXECUTE ON FUNCTION") != 6 ||
		strings.Contains(grants, "GRANT SELECT") ||
		strings.Contains(grants, "GRANT INSERT") ||
		strings.Contains(grants, "GRANT UPDATE") ||
		strings.Contains(grants, "GRANT DELETE") {
		t.Fatal("IAM atomic routine ACL is not execute-only")
	}
}
