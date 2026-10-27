//go:build integration

package postgres

const insertIAMAtomicChallengeFixture = `
INSERT INTO iam_auth_challenges (
    id, user_id, session_id, purpose, proof_kind, hash_profile,
    hmac_key_id, challenge_hmac, context_sha256, step_up_purpose_sha256,
    action_policy_version, state, attempt_count, revision, issued_at,
    expires_at, purge_after, created_at, updated_at
) VALUES (
    '99000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003',
    '81000000-0000-4000-8000-000000000001',
    'reauth_v1', 'password_v1', 'hmac_sha256_v1',
    'iam-challenge-key-v1',
    pg_catalog.decode(pg_catalog.repeat('d1', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('d2', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('d3', 32), 'hex'), 1,
    'pending', 0, 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp() + INTERVAL '5 minutes',
    pg_catalog.transaction_timestamp() + INTERVAL '24 hours 5 minutes',
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
);`

const insertIAMAtomicRefreshFixture = `
INSERT INTO iam_refresh_families (
    id, user_id, session_id, state, rotation_sequence, revision,
    created_at, updated_at, absolute_expires_at, retain_until
) VALUES (
    '92000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002',
    '81000000-0000-4000-8000-000000000002', 'active', 1, 1,
    pg_catalog.transaction_timestamp() - INTERVAL '1 hour',
    pg_catalog.transaction_timestamp() - INTERVAL '1 hour',
    pg_catalog.transaction_timestamp() - INTERVAL '1 hour' + INTERVAL '90 days',
    pg_catalog.transaction_timestamp() - INTERVAL '1 hour' + INTERVAL '91 days'
);
INSERT INTO iam_refresh_tokens (
    id, family_id, family_absolute_expires_at, parent_id, parent_sequence,
    selector, hash_key_id, secret_hmac, sequence, auth_time, acr,
    amr_password, amr_webauthn, amr_totp, amr_recovery,
    step_up_purpose_sha256, state, revision, issued_at, updated_at,
    idle_expires_at, retain_until
)
SELECT '93000000-0000-4000-8000-000000000001', family.id,
       family.absolute_expires_at, NULL, NULL,
       pg_catalog.decode(pg_catalog.repeat('d4', 16), 'hex'),
       'refresh-hmac-v1',
       pg_catalog.decode(pg_catalog.repeat('d5', 32), 'hex'), 1,
       pg_catalog.transaction_timestamp() - INTERVAL '1 hour',
       'login', true, false, false, false, NULL, 'active', 1,
       pg_catalog.transaction_timestamp() - INTERVAL '30 minutes',
       pg_catalog.transaction_timestamp() - INTERVAL '30 minutes',
       pg_catalog.transaction_timestamp() - INTERVAL '30 minutes' + INTERVAL '30 days',
       family.retain_until
FROM iam_refresh_families AS family
WHERE family.id = '92000000-0000-4000-8000-000000000002';`
