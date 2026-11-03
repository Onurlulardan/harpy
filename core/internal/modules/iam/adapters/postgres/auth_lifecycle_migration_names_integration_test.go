//go:build integration

package postgres

const loginTargetKeyContractMigration = "20260725124500_repair_iam_login_target_key_contract"

var iamAtomicRoutineMigrationNames = []string{
	"20260716012950_create_iam_refresh_rotate_v1",
	"20260716012951_create_iam_one_time_consume_v1",
	"20260716012952_create_iam_invitation_accept_v1",
	"20260716012953_create_iam_challenge_resolve_v1",
	"20260716012954_create_iam_recovery_consume_v1",
	"20260716012955_create_iam_portal_setup_complete_v1",
	"20260716012956_grant_iam_atomic_routines_v1",
}

var iamAuthRuntimeMigrationNames = []string{
	"20260716013300_create_iam_login_credential_reader_v1",
	"20260716013310_create_iam_password_use_v1",
	"20260716013320_create_iam_session_issue_v1",
	"20260716013330_create_iam_refresh_credential_reader_v1",
	"20260716013349_create_iam_refresh_decide_v2",
	"20260716013350_create_iam_refresh_reuse_v2",
	"20260716013351_create_iam_refresh_expire_v2",
	"20260716013352_create_iam_refresh_rotate_active_v2",
	"20260716013353_create_iam_refresh_rotate_v2",
	"20260716013354_grant_iam_auth_runtime_v2",
	"20260716013355_create_iam_refresh_locator_source_sha256_v1",
	"20260716013356_create_iam_refresh_credential_reader_v2",
	"20260716013357_create_iam_refresh_rotate_v3",
	"20260716013358_grant_iam_auth_runtime_v3",
	"20260716013359_create_iam_reauthentication_digests_v1",
	"20260716014200_create_iam_reauthentication_start_v1",
	"20260716014201_create_iam_reauthentication_credential_v1",
	"20260716014202_create_iam_reauthentication_reject_v1",
	"20260716014203_create_iam_reauthentication_proof_apply_v1",
	"20260716014204_create_iam_reauthentication_rotate_active_v1",
	"20260716014205_create_iam_reauthentication_complete_v1",
	"20260716014206_create_iam_session_revoke_v1",
	"20260716014207_create_iam_recovery_credential_v1",
	"20260716014208_create_iam_recovery_consume_v2",
	"20260716014209_grant_iam_auth_runtime_v4",
}

var iamAuthorizationRuntimeMigrationNames = []string{
	"20260716013424_create_iam_authorization_evidence_read_v1",
	"20260716013425_grant_iam_authorization_evidence_read_v1",
	"20260716020057_create_iam_authorization_evidence_read_v2",
	"20260716020058_grant_iam_authorization_evidence_read_v2",
	"20260716020100_create_iam_authorization_evidence_read_v3",
	"20260716020101_grant_iam_authorization_evidence_read_v3",
}

var iamAutomationMigrationNames = []string{
	"20260716013400_create_iam_service_principals",
	"20260716013410_create_iam_capability_grants",
	"20260716013420_guard_iam_capability_grants_v1",
	"20260716013421_guard_iam_capability_grant_terminal_v1",
	"20260716013422_create_iam_capability_grant_read_v1",
	"20260716013423_grant_iam_capability_grant_read_v1",
	"20260716013600_create_iam_service_principal_actor_assert_v1",
	"20260716013601_create_iam_service_principal_create_v1",
	"20260716013602_create_iam_service_principal_state_v1",
	"20260716013603_create_iam_capability_grant_write_v1",
	"20260716013604_create_iam_capability_grant_revoke_v1",
	"20260716013605_grant_iam_service_principal_commands_v1",
	"20260716015600_expand_iam_rbac_change_audit_automation_v1",
}

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

var iamLifecycleStartV2QualificationMigrations = []string{
	"20260725124600_qualify_iam_password_reset_start_v2",
	"20260725124601_qualify_iam_email_verification_start_v2",
	"20260725124602_qualify_iam_portal_setup_start_v2",
}

var iamLifecycleKeyContractMigrations = []string{
	"20260725124700_repair_iam_one_time_key_contract",
	"20260725124701_repair_iam_password_reset_complete_key_contract",
}
