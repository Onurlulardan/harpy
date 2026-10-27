package postgres

import (
	"strings"
	"testing"
)

func TestIAMLifecycleRoutineSecurityContracts(t *testing.T) {
	for _, name := range iamLifecycleMigrationNames[1:] {
		source := string(readIAMIdentityMigration(t, "fresh", name))
		if !strings.Contains(source, "SECURITY DEFINER") ||
			!strings.Contains(source, "SET search_path FROM CURRENT") ||
			!strings.Contains(source, "REVOKE ALL ON FUNCTION") ||
			!strings.Contains(source, "FROM PUBLIC;") {
			t.Fatalf("%s lacks a fail-closed routine boundary", name)
		}
	}
}

func TestIAMLifecycleCredentialCommandsArePurposeBound(t *testing.T) {
	checks := map[string][]string{
		iamLifecycleMigrationNames[3]: {
			"purpose = 'email_verification_v1'", "token.secret_hmac = p_secret_hmac",
			"token.binding_sha256 = p_binding_sha256", "state = 'consumed'",
		},
		iamLifecycleMigrationNames[8]: {
			"'password_reset_v1'", "p_secret_hmac", "p_binding_sha256",
		},
		iamLifecycleMigrationNames[10]: {
			"purpose = 'password_reset_v1'", "token.secret_hmac = p_secret_hmac",
			"token.binding_sha256 = p_binding_sha256", "state = 'consumed'",
		},
		iamLifecycleMigrationNames[11]: {
			"purpose = 'portal_setup_v1'", "token.secret_hmac = p_secret_hmac",
			"token.binding_sha256 = p_binding_sha256", "state = 'consumed'",
		},
	}
	for name, markers := range checks {
		source := string(readIAMIdentityMigration(t, "fresh", name))
		for _, marker := range markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", name, marker)
			}
		}
	}
}

func TestIAMLifecycleSecurityChangesBumpAndRevokeAtomically(t *testing.T) {
	for _, index := range []int{3, 4, 5, 10, 11, 12, 13, 15, 16, 18} {
		name := iamLifecycleMigrationNames[index]
		source := string(readIAMIdentityMigration(t, "fresh", name))
		if !strings.Contains(source, "iam_lifecycle_record_v1(") {
			t.Fatalf("%s does not record its atomic security effect", name)
		}
	}
	for _, index := range []int{10, 13} {
		name := iamLifecycleMigrationNames[index]
		source := string(readIAMIdentityMigration(t, "fresh", name))
		for _, marker := range []string{
			"UPDATE iam_sessions", "UPDATE iam_refresh_families",
			"UPDATE iam_refresh_tokens",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", name, marker)
			}
		}
	}
}
