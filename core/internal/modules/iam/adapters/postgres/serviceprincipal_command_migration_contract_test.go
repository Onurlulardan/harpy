package postgres

import (
	"strings"
	"testing"
)

func TestIAMServicePrincipalCommandsFreezeCASAndStepUp(t *testing.T) {
	tests := []struct {
		name    string
		markers []string
	}{
		{"20260716013600_create_iam_service_principal_actor_assert_v1", []string{
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"p_auth_time < v_now - INTERVAL '10 minutes'",
			"iam_user_has_permission_v1", "FOR UPDATE",
		}},
		{"20260716013601_create_iam_service_principal_create_v1", []string{
			"iam.service_principal.create", "iam_authorization_operation_receipts",
			"iam_authorization_clock", "authorization_version",
		}},
		{"20260716013602_create_iam_service_principal_state_v1", []string{
			"p_expected_revision", "iam_capability_grants", "active capability grants",
		}},
		{"20260716013603_create_iam_capability_grant_write_v1", []string{
			"v_new_revision := p_expected_revision + 1",
			"iam_capability_grant_revisions", "manifest_sha256",
		}},
		{"20260716013604_create_iam_capability_grant_revoke_v1", []string{
			"disposition", "'revoked'", "iam.capability_grant.revoke",
		}},
	}
	for _, test := range tests {
		document := string(readIAMIdentityMigration(t, "fresh", test.name))
		for _, marker := range test.markers {
			if !strings.Contains(document, marker) {
				t.Fatalf("%s lacks %q", test.name, marker)
			}
		}
	}
}
