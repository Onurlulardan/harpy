//go:build integration

package postgres

const insertIAMAuthenticators = `
INSERT INTO iam_authenticators (
    id, user_id, kind, display_label, state, revision,
    created_at, updated_at, activated_at
) VALUES
    ('84000000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000003',
     'webauthn_v1', 'Primary passkey', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z'),
    ('84000000-0000-4000-8000-000000000002',
     '51000000-0000-4000-8000-000000000002',
     'totp_v1', 'Authenticator app', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:05:00Z',
     '2026-07-16T00:05:00Z'),
    ('84000000-0000-4000-8000-000000000003',
     '51000000-0000-4000-8000-000000000001',
     'recovery_set_v1', 'Recovery codes', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z');`

const insertIAMAuthenticatorSubtypes = `
INSERT INTO iam_webauthn_credentials (
    authenticator_id, user_id, authenticator_kind, credential_id,
    public_key_cose, rp_id, rp_id_sha256, origin_policy_version,
    aaguid, sign_count, backup_eligible, backup_state, discoverable,
    uv_required, revision, created_at, updated_at
) VALUES (
    '84000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003', 'webauthn_v1',
    pg_catalog.decode(pg_catalog.repeat('81', 32), 'hex'),
    pg_catalog.decode('a5010203262001215820', 'hex'),
    'auth.harpy.example',
    pg_catalog.decode(pg_catalog.repeat('82', 32), 'hex'), 1,
    '87000000-0000-4000-8000-000000000001', 0,
    true, false, true, true, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
);

INSERT INTO iam_totp_authenticators (
    authenticator_id, user_id, authenticator_kind, profile, cipher_suite,
    kms_key_id, nonce, secret_ciphertext, authentication_tag, aad_sha256,
    last_accepted_time_step, verified_at, enrollment_expires_at,
    revision, created_at, updated_at
) VALUES (
    '84000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002', 'totp_v1',
    'totp_sha256_6_30_v1', 'aes_256_gcm_v1', 'iam-totp-key-v1',
    pg_catalog.decode(pg_catalog.repeat('83', 12), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('84', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('85', 16), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('86', 32), 'hex'),
    NULL, '2026-07-16T00:05:00Z', '2026-07-16T00:10:00Z', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:05:00Z'
);

INSERT INTO iam_recovery_codes (
    id, authenticator_id, user_id, authenticator_kind, position,
    hash_profile, hmac_key_id, code_hmac, state, revision,
    created_at, updated_at
) VALUES (
    '85000000-0000-4000-8000-000000000001',
    '84000000-0000-4000-8000-000000000003',
    '51000000-0000-4000-8000-000000000001', 'recovery_set_v1', 1,
    'hmac_sha256_v1', 'iam-recovery-key-v1',
    pg_catalog.decode(pg_catalog.repeat('87', 32), 'hex'),
    'active', 1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
);`
