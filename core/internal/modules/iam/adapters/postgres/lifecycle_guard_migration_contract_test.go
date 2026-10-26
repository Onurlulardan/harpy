package postgres

import (
	"strings"
	"testing"
)

func TestIAMLifecycleGuardMigrationContracts(t *testing.T) {
	tests := []struct {
		name      string
		functions int
		triggers  int
		markers   []string
	}{
		{iamLifecycleGuardMigrationNames[0], 1, 20, []string{
			"CREATE FUNCTION iam_guard_revision_step_v1()",
			"NEW.revision IS DISTINCT FROM OLD.revision + 1",
			"iam_refresh_tokens_revision_guard",
		}},
		{iamLifecycleGuardMigrationNames[1], 1, 10, []string{
			"CREATE FUNCTION iam_guard_immutable_columns_v1()",
			"pg_catalog.to_jsonb(OLD)",
			"iam_portal_access_setups_immutable_guard",
		}},
		{iamLifecycleGuardMigrationNames[2], 0, 5, []string{
			"iam_sessions_immutable_guard",
			"iam_totp_authenticators_immutable_guard",
			"'secret_ciphertext'",
		}},
		{iamLifecycleGuardMigrationNames[3], 0, 5, []string{
			"iam_auth_challenges_immutable_guard",
			"iam_refresh_tokens_immutable_guard",
			"'parent_sequence'", "'secret_hmac'", "'issued_at'",
			"'identifier_revision'", "'portal_setup_revision'",
		}},
		{iamLifecycleGuardMigrationNames[4], 1, 1, []string{
			"CREATE FUNCTION iam_guard_user_security_versions_v1()",
			"iam_users_security_version_guard",
		}},
		{iamLifecycleGuardMigrationNames[5], 1, 7, []string{
			"CREATE FUNCTION iam_guard_state_transition_v1()",
			"'active>consumed'", "'pending>accepted'",
			"iam_invitations_state_transition_guard",
		}},
		{iamLifecycleGuardMigrationNames[6], 1, 6, []string{
			"CREATE FUNCTION iam_guard_monotonic_bigint_v1()",
			"iam_webauthn_credentials_sign_count_guard",
			"iam_totp_authenticators_last_step_guard",
			"iam_refresh_families_rotation_guard",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := string(readIAMIdentityMigration(t, "fresh", test.name))
			if strings.Count(source, "CREATE FUNCTION ") != test.functions ||
				strings.Count(source, "CREATE TRIGGER ") != test.triggers {
				t.Fatal("IAM lifecycle function/trigger count differs")
			}
			for _, marker := range test.markers {
				if !strings.Contains(source, marker) {
					t.Fatalf("IAM lifecycle guard lacks %q", marker)
				}
			}
		})
	}
}
