//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAuthenticatorChallengeConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAuthenticators)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAuthenticatorSubtypes)
		assertIAMAuthSecretColumnTypes(t, ctx, database, tenant.schema)
		assertIAMAuthenticatorSubtypeConstraints(t, ctx, database, tenant.schema)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAuthChallenges)
		assertIAMAuthChallengeConstraints(t, ctx, database, tenant.schema)
	}
}

func assertIAMAuthenticatorSubtypeConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	tests := []struct {
		statement  string
		code       string
		constraint string
	}{
		{`UPDATE iam_webauthn_credentials
SET user_id = '51000000-0000-4000-8000-000000000001'
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`,
			"23503", "iam_webauthn_credentials_authenticator_fk"},
		{`UPDATE iam_totp_authenticators
SET user_id = '51000000-0000-4000-8000-000000000001'
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`,
			"23503", "iam_totp_authenticators_authenticator_fk"},
		{`UPDATE iam_webauthn_credentials SET rp_id = 'auth..harpy.example'
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`,
			"23514", "iam_webauthn_credentials_rp_id_check"},
		{`UPDATE iam_webauthn_credentials SET origin_policy_version = 0
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`,
			"23514", "iam_webauthn_credentials_binding_check"},
		{`UPDATE iam_webauthn_credentials SET sign_count = -1
WHERE authenticator_id = '84000000-0000-4000-8000-000000000001'`,
			"23514", "iam_webauthn_credentials_counter_check"},
		{`UPDATE iam_totp_authenticators SET profile = 'totp_sha1_6_30_v1'
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`,
			"23514", "iam_totp_authenticators_profile_check"},
		{`UPDATE iam_totp_authenticators SET cipher_suite = 'aes_128_gcm_v1'
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`,
			"23514", "iam_totp_authenticators_cipher_suite_check"},
		{`UPDATE iam_totp_authenticators
SET nonce = pg_catalog.decode(pg_catalog.repeat('aa', 11), 'hex')
WHERE authenticator_id = '84000000-0000-4000-8000-000000000002'`,
			"23514", "iam_totp_authenticators_secret_envelope_check"},
		{`INSERT INTO iam_authenticators (
    id, user_id, kind, display_label, state, revision, created_at, updated_at
) VALUES (
    '84000000-0000-4000-8000-000000000099',
    '51000000-0000-4000-8000-000000000001',
    'totp_v1', 'Second app', 'pending', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
);
INSERT INTO iam_totp_authenticators (
    authenticator_id, user_id, authenticator_kind, profile, cipher_suite,
    kms_key_id, nonce, secret_ciphertext, authentication_tag, aad_sha256,
    enrollment_expires_at, revision, created_at, updated_at
) VALUES (
    '84000000-0000-4000-8000-000000000099',
    '51000000-0000-4000-8000-000000000001', 'totp_v1',
    'totp_sha256_6_30_v1', 'aes_256_gcm_v1', 'iam-totp-key-v1',
    pg_catalog.decode(pg_catalog.repeat('83', 12), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('a5', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('a6', 16), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('a7', 32), 'hex'),
    '2026-07-16T00:10:00Z', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
)`, "23505", "iam_totp_authenticators_nonce_key"},
		{`UPDATE iam_recovery_codes
SET code_hmac = pg_catalog.decode(pg_catalog.repeat('bb', 31), 'hex')
WHERE id = '85000000-0000-4000-8000-000000000001'`,
			"23514", "iam_recovery_codes_hmac_check"},
		{`UPDATE iam_recovery_codes
SET state = 'consumed', consumed_at = updated_at, purge_after = NULL
WHERE id = '85000000-0000-4000-8000-000000000001'`,
			"23514", "iam_recovery_codes_lifecycle_check"},
	}
	for _, test := range tests {
		t.Run(schema+"/"+test.constraint, func(t *testing.T) {
			err := rejectIAMTenant(t, ctx, database, schema, test.statement)
			assertIAMConstraint(t, err, test.code, test.constraint)
		})
	}
}

func assertIAMAuthSecretColumnTypes(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	const query = `
WITH expected(table_name, column_name) AS (VALUES
    ('iam_webauthn_credentials', 'credential_id'),
    ('iam_webauthn_credentials', 'public_key_cose'),
    ('iam_webauthn_credentials', 'rp_id_sha256'),
    ('iam_totp_authenticators', 'nonce'),
    ('iam_totp_authenticators', 'secret_ciphertext'),
    ('iam_totp_authenticators', 'authentication_tag'),
    ('iam_totp_authenticators', 'aad_sha256'),
    ('iam_recovery_codes', 'code_hmac'),
    ('iam_auth_challenges', 'challenge_hmac'),
    ('iam_auth_challenges', 'context_sha256'),
    ('iam_auth_challenges', 'step_up_purpose_sha256'),
    ('iam_auth_challenges', 'rp_id_sha256'),
    ('iam_auth_challenges', 'origin_sha256')
)
SELECT pg_catalog.count(c.column_name),
       pg_catalog.count(c.column_name) FILTER (WHERE c.data_type = 'bytea')
FROM expected AS e
LEFT JOIN information_schema.columns AS c
  ON c.table_schema = $1 AND c.table_name = e.table_name
 AND c.column_name = e.column_name`
	var present, binary int
	if err := transaction.QueryRowContext(ctx, query, schema).Scan(&present, &binary); err != nil {
		t.Fatalf("inspect %s IAM secret columns: %v", schema, err)
	}
	if present != 13 || binary != 13 {
		t.Fatalf("%s IAM protected columns present/binary = %d/%d, want 13/13", schema, present, binary)
	}
	const forbiddenQuery = `
SELECT pg_catalog.count(*)
FROM information_schema.columns
WHERE table_schema = $1
  AND table_name IN (
      'iam_webauthn_credentials', 'iam_totp_authenticators',
      'iam_recovery_codes', 'iam_auth_challenges'
  )
  AND column_name IN (
      'secret', 'plaintext_secret', 'totp_secret', 'recovery_code',
      'challenge', 'challenge_plaintext'
  )`
	var forbidden int
	if err := transaction.QueryRowContext(ctx, forbiddenQuery, schema).Scan(&forbidden); err != nil {
		t.Fatalf("inspect %s IAM plaintext columns: %v", schema, err)
	}
	if forbidden != 0 {
		t.Fatalf("%s IAM plaintext secret columns = %d, want 0", schema, forbidden)
	}
}
