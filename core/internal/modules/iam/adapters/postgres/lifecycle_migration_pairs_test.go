package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamLifecycleMigrationNames = []string{
	"20260716014500_create_iam_lifecycle_audit",
	"20260716014501_create_iam_lifecycle_receipt_v1",
	"20260716014502_create_iam_identifier_add_v1",
	"20260716014503_create_iam_identifier_verify_v1",
	"20260716014504_create_iam_identifier_tombstone_v1",
	"20260716014505_create_iam_identifier_retire_v1",
	"20260716014506_create_iam_one_time_issue_v1",
	"20260716014507_create_iam_email_verification_start_v1",
	"20260716014508_create_iam_password_reset_start_v1",
	"20260716014509_create_iam_portal_setup_start_v1",
	"20260716014510_create_iam_password_reset_complete_v1",
	"20260716014511_create_iam_portal_setup_complete_v2",
	"20260716014512_create_iam_portal_approval_decide_v1",
	"20260716014513_create_iam_password_change_v1",
	"20260716014514_create_iam_authenticator_list_v1",
	"20260716014515_create_iam_authenticator_suspend_v1",
	"20260716014516_create_iam_authenticator_retire_v1",
	"20260716014517_create_iam_webauthn_enrollment_start_v1",
	"20260716014518_create_iam_webauthn_enrollment_complete_v1",
}

var iamLifecycleRuntimeMigrationNames = []string{
	"20260716014519_create_iam_identifier_verify_v2",
	"20260716014520_expand_iam_totp_kms_key_id_v2",
	"20260716014521_create_iam_totp_enrollment_start_v1",
	"20260716014522_create_iam_totp_enrollment_read_v1",
	"20260716014523_create_iam_totp_enrollment_complete_v1",
	"20260716014524_create_iam_recovery_codes_issue_v1",
	"20260716014525_create_iam_recovery_codes_list_v1",
	"20260716014526_create_iam_session_list_v1",
	"20260716014527_create_iam_session_revoke_one_v1",
	"20260716014528_create_iam_session_revoke_all_v1",
	"20260716014529_create_iam_lifecycle_event_sequence_v1",
	"20260716014530_grant_iam_lifecycle_runtime_v1",
	"20260716014531_grant_iam_lifecycle_security_runtime_v1",
	"20260716014532_create_iam_password_credential_read_v1",
	"20260716014533_create_iam_authenticator_suspend_v2",
	"20260716014534_create_iam_authenticator_retire_v2",
	"20260716014535_create_iam_lifecycle_credential_lookup_v1",
	"20260716014536_grant_iam_lifecycle_credential_lookup_v1",
	"20260716014537_create_iam_lifecycle_delivery_artifacts",
	"20260716014538_create_iam_lifecycle_delivery_artifact_insert_v1",
	"20260716014539_create_iam_lifecycle_delivery_targets_v1",
	"20260716014540_create_iam_email_verification_start_v2",
	"20260716014541_create_iam_password_reset_start_v2",
	"20260716014542_create_iam_portal_setup_start_v2",
	"20260716014543_grant_iam_lifecycle_delivery_v1",
	"20260716014544_create_iam_lifecycle_delivery_worker_claims",
	"20260716014545_create_iam_lifecycle_delivery_worker_receipts",
}

func TestIAMLifecycleRuntimeMigrationsStayPairedAndFrozen(t *testing.T) {
	if len(iamLifecycleRuntimeMigrationNames) != 27 {
		t.Fatalf("lifecycle runtime migration count=%d want 27",
			len(iamLifecycleRuntimeMigrationNames))
	}
	for _, name := range iamLifecycleRuntimeMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted IAM lifecycle runtime differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM lifecycle runtime migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM lifecycle runtime migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"CREATE TABLE IF NOT EXISTS", "CREATE FUNCTION IF NOT EXISTS",
				"CREATE OR REPLACE", "-- +GOOSE DOWN", "CREATE TEMP",
				"CREATE TEMPORARY", "UNLOGGED", "PG_TEMP", "DISABLE TRIGGER",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("IAM lifecycle runtime contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMLifecycleMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamLifecycleMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted IAM lifecycle differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM lifecycle migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM lifecycle migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"CREATE TABLE IF NOT EXISTS", "CREATE FUNCTION IF NOT EXISTS",
				"CREATE OR REPLACE", "-- +GOOSE DOWN",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"DISABLE TRIGGER", "ON DELETE CASCADE", "\nGRANT ",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("IAM lifecycle contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMLifecycleAuditIsRedacted(t *testing.T) {
	source := strings.ToLower(string(readIAMIdentityMigration(
		t, "fresh", iamLifecycleMigrationNames[0],
	)))
	for _, forbidden := range []string{
		"secret_hmac", "password", "ciphertext", "authentication_tag",
		"public_key_cose", "credential_id", "challenge_hmac", "code_hmac",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("lifecycle audit contains secret-bearing field %q", forbidden)
		}
	}
}
