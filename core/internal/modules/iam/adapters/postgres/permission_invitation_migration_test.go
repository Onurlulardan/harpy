package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamPermissionInvitationMigrationNames = []string{
	"20260716013430_create_iam_permission_request_groups",
	"20260716013431_create_iam_permission_requests",
	"20260716013432_guard_iam_permission_requests_v1",
	"20260716013433_create_iam_user_has_permission_v1",
	"20260716013434_create_iam_permission_request_group_v1",
	"20260716013435_create_iam_permission_request_review_v1",
	"20260716013436_create_iam_permission_request_cancel_v1",
	"20260716013437_create_iam_permission_request_expire_v1",
	"20260716013440_create_iam_permission_request_grant_v1",
	"20260716013441_create_iam_permission_request_bump_users_v1",
	"20260716013442_create_iam_permission_request_decide_v1",
	"20260716013443_grant_iam_permission_request_routines_v1",
	"20260716013444_create_iam_permission_request_read_v1",
	"20260716013445_create_iam_permission_request_list_v1",
	"20260716013446_grant_iam_permission_request_reads_v1",
	"20260716013447_create_iam_invitation_manager_v1",
	"20260716013448_create_iam_invitation_create_v1",
	"20260716013449_create_iam_invitation_read_v1",
	"20260716013450_create_iam_invitation_list_v1",
	"20260716013451_create_iam_invitation_revoke_v1",
	"20260716013452_create_iam_invitation_expire_v1",
	"20260716013453_create_iam_invitation_grant_role_v1",
	"20260716013454_create_iam_invitation_accept_v2",
	"20260716013455_grant_iam_invitation_routines_v1",
	"20260716015000_create_iam_permission_request_decision_audit",
	"20260716015001_guard_iam_permission_request_authority_v1",
	"20260716015002_create_iam_permission_request_audit_write_v1",
	"20260716015003_grant_iam_permission_request_audit_v1",
	"20260716015004_create_iam_permission_request_transition_audit",
	"20260716015005_create_iam_permission_request_transition_audit_write_v1",
	"20260716015006_grant_iam_permission_request_transition_audit_v1",
	"20260716015200_create_iam_invitation_delivery_artifacts",
	"20260716015201_create_iam_invitation_receipts_and_audit",
	"20260716015202_create_iam_invitation_create_v2",
	"20260716015203_create_iam_invitation_revoke_v2",
	"20260716015204_create_iam_invitation_expire_v2",
	"20260716015205_create_iam_invitation_accept_v3",
	"20260716015206_create_iam_invitation_audit_write_v1",
	"20260716015207_grant_iam_invitation_runtime_v2",
	"20260716015208_create_iam_invitation_role_lookup_v1",
	"20260716015209_create_iam_invitation_credential_lookup_v1",
	"20260716015210_create_iam_invitation_read_v2",
	"20260716015211_create_iam_invitation_list_v2",
	"20260716015212_create_iam_invitation_register_accept_v1",
	"20260716015213_grant_iam_invitation_http_v1",
	"20260801011000_create_iam_permission_request_resource_catalog_v1",
	"20260801011001_create_iam_permission_request_action_catalog_v1",
	"20260801011002_grant_iam_permission_request_catalog_v1",
}

func TestIAMPermissionInvitationMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamPermissionInvitationMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migration content differs")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("permission/invitation migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMInvitationPlatformMigrationStaysPairedAndClosed(t *testing.T) {
	const name = "20260716015300_create_iam_identity_directory_upsert_v2"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	adopted := readIAMPlatformAuthMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("invitation platform fresh and v1_adopted migration differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("invitation platform migration is invalid: %v", err)
	}
}

func TestIAMPermissionInvitationSecurityContract(t *testing.T) {
	decision := string(readIAMIdentityMigration(
		t, "fresh", "20260716013442_create_iam_permission_request_decide_v1",
	))
	for _, marker := range []string{
		"urn:harpy:acr:step-up", "FOR UPDATE",
		"iam_authorization_operation_receipts", "iam_permission_request_grant_v1",
	} {
		if !strings.Contains(decision, marker) {
			t.Fatalf("permission decision lacks %q", marker)
		}
	}
	accept := string(readIAMIdentityMigration(
		t, "fresh", "20260716013454_create_iam_invitation_accept_v2",
	))
	for _, marker := range []string{
		"v_token_hmac IS DISTINCT FROM p_secret_hmac",
		"iam_invitation_grant_role_v1", "iam_authorization_operation_receipts",
		"authorization_version = authorization_version + 1",
	} {
		if !strings.Contains(accept, marker) {
			t.Fatalf("invitation acceptance lacks %q", marker)
		}
	}
}
