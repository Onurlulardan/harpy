package postgres

import (
	"strings"
	"testing"
)

func TestIAMOrganizationOwnerMigrationContracts(t *testing.T) {
	tests := []struct {
		name    string
		markers []string
	}{
		{"20260802232000_create_iam_identity_organization_owners_replace_v1", []string{
			"CREATE FUNCTION iam_identity_organization_owners_replace_v1(",
			"membership_kind = 'member'", "membership_kind = 'owner'",
			"membership.state = 'tombstoned'", "iam_identity_user_bump_v1",
			"exact active owner set",
		}},
		{"20260802232001_create_iam_identity_organization_create_v3", []string{
			"CREATE FUNCTION iam_identity_organization_create_v3(",
			"iam_identity_organization_owners_replace_v1(",
			"iam_identity_record_v1(",
		}},
		{"20260802232002_create_iam_identity_organization_update_v3", []string{
			"CREATE FUNCTION iam_identity_organization_update_v3(",
			"p_replace_owners boolean", "requires an active owner",
			"iam_identity_organization_owners_replace_v1(",
		}},
		{"20260802232003_grant_iam_identity_organization_owner_v3", []string{
			"REVOKE EXECUTE ON FUNCTION iam_identity_organization_create_v1(",
			"REVOKE EXECUTE ON FUNCTION iam_identity_organization_create_v2(",
			"GRANT EXECUTE ON FUNCTION iam_identity_organization_create_v3(",
			"GRANT EXECUTE ON FUNCTION iam_identity_organization_update_v3(",
		}},
		{"20260802232004_create_iam_identity_organization_delete_v2", []string{
			"CREATE FUNCTION iam_identity_organization_delete_v2(",
			"v_status <> 'suspended'", "state <> 'revoked'",
			"SET state = 'tombstoned'", "iam_identity_user_bump_v1",
		}},
		{"20260802232005_grant_iam_identity_organization_delete_v2", []string{
			"REVOKE EXECUTE ON FUNCTION iam_identity_organization_delete_v1(",
			"GRANT EXECUTE ON FUNCTION iam_identity_organization_delete_v2(",
		}},
	}
	for _, test := range tests {
		source := string(readIAMIdentityMigration(t, "fresh", test.name))
		for _, marker := range test.markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", test.name, marker)
			}
		}
	}
}
