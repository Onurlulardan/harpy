package postgres

import (
	"strings"
	"testing"
)

func TestIAMAuthenticationMigrationContracts(t *testing.T) {
	tests := []struct {
		name        string
		foreignKeys int
		markers     []string
	}{
		{name: iamAuthenticationMigrationNames[0], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_authenticators (",
			"display_label text NOT NULL",
			"UNIQUE (id, user_id, kind)",
			"REFERENCES iam_users (id)",
			"kind IN ('webauthn_v1', 'totp_v1', 'recovery_set_v1')",
			"'compromised'", "'renewal_required', 'retired'",
			"state = 'compromised' AND compromised_at IS NOT NULL",
			"state = 'renewal_required' AND activated_at IS NOT NULL",
			"iam_authenticators_one_live_totp_uidx",
			"iam_authenticators_one_live_recovery_set_uidx",
		}},
		{name: iamAuthenticationMigrationNames[1], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_webauthn_credentials (",
			"PRIMARY KEY (authenticator_id)", "UNIQUE (credential_id)",
			"REFERENCES iam_authenticators (id, user_id, kind)",
			"CHECK (authenticator_kind = 'webauthn_v1')",
			"pg_catalog.octet_length(credential_id) BETWEEN 1 AND 1024",
			"pg_catalog.octet_length(public_key_cose) BETWEEN 1 AND 4096",
			"rp_id text COLLATE \"C\" NOT NULL",
			"pg_catalog.octet_length(rp_id_sha256) = 32",
			"origin_policy_version > 0", "aaguid uuid NOT NULL",
			"CHECK (sign_count >= 0)",
			"CHECK (NOT backup_state OR backup_eligible)",
			"discoverable boolean NOT NULL", "uv_required boolean NOT NULL",
		}},
		{name: iamAuthenticationMigrationNames[2], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_totp_authenticators (",
			"REFERENCES iam_authenticators (id, user_id, kind)",
			"CHECK (authenticator_kind = 'totp_v1')",
			"CHECK (profile = 'totp_sha256_6_30_v1')",
			"CHECK (cipher_suite = 'aes_256_gcm_v1')",
			"pg_catalog.octet_length(nonce) = 12",
			"pg_catalog.octet_length(secret_ciphertext) = 32",
			"pg_catalog.octet_length(authentication_tag) = 16",
			"pg_catalog.octet_length(aad_sha256) = 32",
			"enrollment_expires_at = created_at + INTERVAL '10 minutes'",
			"verified_at <= enrollment_expires_at",
		}},
		{name: iamAuthenticationMigrationNames[3], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_recovery_codes (",
			"UNIQUE (authenticator_id, position)",
			"UNIQUE (hmac_key_id, code_hmac)",
			"REFERENCES iam_authenticators (id, user_id, kind)",
			"CHECK (authenticator_kind = 'recovery_set_v1')",
			"CHECK (position BETWEEN 1 AND 32)",
			"CHECK (hash_profile = 'hmac_sha256_v1')",
			"CHECK (pg_catalog.octet_length(code_hmac) = 32)",
			"state IN ('active', 'consumed', 'revoked')",
			"purge_after = consumed_at + INTERVAL '30 days'",
			"purge_after = revoked_at + INTERVAL '30 days'",
		}},
		{name: iamAuthenticationMigrationNames[4], foreignKeys: 3, markers: []string{
			"CREATE TABLE iam_auth_challenges (",
			"REFERENCES iam_users (id)",
			"REFERENCES iam_sessions (id, user_id)",
			"REFERENCES iam_authenticators (id, user_id, kind)",
			"'webauthn_register_v1', 'webauthn_authenticate_v1'",
			"'totp_enroll_v1', 'reauth_v1'",
			"proof_kind IN ('password_v1', 'webauthn_v1', 'totp_v1')",
			"proof_kind = 'password_v1' AND",
			"pg_catalog.octet_length(step_up_purpose_sha256) = 32",
			"pg_catalog.octet_length(challenge_hmac) = 32",
			"pg_catalog.octet_length(context_sha256) = 32",
			"attempt_count BETWEEN 0 AND 5",
			"expires_at <= issued_at + INTERVAL '5 minutes'",
			"state IN ('pending', 'consumed', 'failed', 'expired', 'cancelled')",
			"purge_after >= GREATEST(",
			"iam_auth_challenges_retention_idx",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := string(readIAMAuthenticationMigration(t, "fresh", test.name))
			assertIAMAuthenticationShape(t, source, test.foreignKeys, test.markers)
		})
	}
}

func assertIAMAuthenticationShape(
	t *testing.T,
	source string,
	foreignKeys int,
	markers []string,
) {
	t.Helper()
	if strings.Count(source, "CREATE TABLE ") != 1 ||
		strings.Count(source, "FOREIGN KEY (") != foreignKeys ||
		strings.Count(source, "ON UPDATE NO ACTION ON DELETE RESTRICT") != foreignKeys {
		t.Fatal("IAM authentication migration relation shape differs")
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM authentication migration lacks exact marker %q", marker)
		}
	}
}
