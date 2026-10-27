package postgres

import (
	"bytes"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamIdentityPlatformMigrationNames = []string{
	"20260716013705_create_iam_identity_directory_upsert_v1",
	"20260716013706_create_iam_identity_directory_tombstone_v1",
	"20260716013707_grant_iam_identity_directory_runtime_v1",
}

var iamIdentityTenantMigrationNames = []string{
	"20260716013700_create_iam_identity_authority_v1",
	"20260716013701_create_iam_identity_profile_assert_v1",
	"20260716013702_create_iam_identity_receipt_v1",
	"20260716013703_create_iam_identity_user_bump_v1",
	"20260716013710_create_iam_identity_user_create_v1",
	"20260716013711_create_iam_identity_user_read_v1",
	"20260716013712_create_iam_identity_user_list_v1",
	"20260716013713_create_iam_identity_user_update_v1",
	"20260716013714_create_iam_identity_user_disable_v1",
	"20260716013715_create_iam_identity_user_delete_v1",
	"20260716013720_create_iam_identity_organization_create_v1",
	"20260716013721_create_iam_identity_organization_read_v1",
	"20260716013722_create_iam_identity_organization_list_v1",
	"20260716013723_create_iam_identity_organization_update_v1",
	"20260716013724_create_iam_identity_organization_disable_v1",
	"20260716013725_create_iam_identity_organization_delete_v1",
	"20260716013730_create_iam_identity_role_create_v1",
	"20260716013731_create_iam_identity_role_read_v1",
	"20260716013732_create_iam_identity_role_list_v1",
	"20260716013733_create_iam_identity_role_update_v1",
	"20260716013734_create_iam_identity_role_retire_v1",
	"20260716013740_create_iam_identity_membership_add_v1",
	"20260716013741_create_iam_identity_membership_read_v1",
	"20260716013742_create_iam_identity_membership_list_v1",
	"20260716013743_create_iam_identity_membership_update_v1",
	"20260716013744_create_iam_identity_membership_delete_v1",
	"20260716013750_create_iam_identity_assignment_grant_v1",
	"20260716013751_create_iam_identity_assignment_read_v1",
	"20260716013752_create_iam_identity_assignment_list_v1",
	"20260716013753_create_iam_identity_assignment_update_v1",
	"20260716013754_create_iam_identity_assignment_revoke_v1",
	"20260716013759_grant_iam_identity_runtime_v1",
	"20260716014305_create_iam_profile_backfill_plans",
	"20260716014700_create_iam_identity_password_enroll_v1",
	"20260716015400_create_iam_identity_change_audit",
	"20260716015401_create_iam_identity_change_audit_write_v1",
	"20260716015402_grant_iam_identity_change_audit_v1",
	"20260716020020_create_iam_identity_profile_assert_v2",
	"20260716020024_create_iam_identity_organization_create_v2",
	"20260716020025_create_iam_identity_organization_update_v2",
	"20260716020051_create_iam_identity_user_relations_replace_v1",
	"20260716020052_grant_iam_identity_user_relations_replace_v1",
	"20260716020053_create_iam_identity_user_read_v2",
	"20260716020054_create_iam_identity_user_list_v2",
	"20260716020055_grant_iam_identity_user_reads_v2",
	"20260802232000_create_iam_identity_organization_owners_replace_v1",
	"20260802232001_create_iam_identity_organization_create_v3",
	"20260802232002_create_iam_identity_organization_update_v3",
	"20260802232003_grant_iam_identity_organization_owner_v3",
	"20260802232004_create_iam_identity_organization_delete_v2",
	"20260802232005_grant_iam_identity_organization_delete_v2",
}

func TestIAMIdentityRuntimeMigrationsStayPairedAndValid(t *testing.T) {
	for _, migration := range iamIdentityPlatformMigrationNames {
		assertIAMIdentityMigrationPair(t, "platform", migration)
	}
	for _, migration := range iamIdentityTenantMigrationNames {
		assertIAMIdentityMigrationPair(t, "tenant", migration)
	}
}

func assertIAMIdentityMigrationPair(t *testing.T, scope, name string) {
	t.Helper()
	var fresh, adopted []byte
	if scope == "platform" {
		fresh = readIAMPlatformAuthMigration(t, "fresh", name)
		adopted = readIAMPlatformAuthMigration(t, "v1_adopted", name)
	} else {
		fresh = readIAMIdentityMigration(t, "fresh", name)
		adopted = readIAMIdentityMigration(t, "v1_adopted", name)
	}
	if !bytes.Equal(fresh, adopted) {
		t.Fatalf("%s identity migration %s differs by origin", scope, name)
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatalf("%s identity migration %s exceeds 200 lines", scope, name)
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect %s identity migration %s: %v", scope, name, err)
	}
}
