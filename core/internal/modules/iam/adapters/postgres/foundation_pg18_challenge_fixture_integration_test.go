//go:build integration

package postgres

const insertIAMAuthChallenges = `
INSERT INTO iam_auth_challenges (
    id, user_id, session_id, authenticator_id, authenticator_kind,
    purpose, proof_kind, hash_profile, hmac_key_id, challenge_hmac,
    context_sha256, origin_policy_version, rp_id_sha256, origin_sha256,
    user_verification_required, state, attempt_count, revision,
    issued_at, expires_at, purge_after, created_at, updated_at
) VALUES (
    '86000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003',
    '81000000-0000-4000-8000-000000000001',
    '84000000-0000-4000-8000-000000000001', 'webauthn_v1',
    'webauthn_register_v1', 'webauthn_v1', 'hmac_sha256_v1',
    'iam-challenge-key-v1',
    pg_catalog.decode(pg_catalog.repeat('91', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('92', 32), 'hex'), 1,
    pg_catalog.decode(pg_catalog.repeat('93', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('94', 32), 'hex'), true,
    'pending', 0, 1, '2026-07-16T00:10:00Z',
    '2026-07-16T00:15:00Z', '2026-07-17T00:15:00Z',
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z'
);

INSERT INTO iam_auth_challenges (
    id, user_id, session_id, authenticator_id, authenticator_kind,
    purpose, proof_kind, hash_profile, hmac_key_id, challenge_hmac,
    context_sha256, state, attempt_count, revision, issued_at, expires_at,
    purge_after, created_at, updated_at
) VALUES (
    '86000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002',
    '81000000-0000-4000-8000-000000000002',
    '84000000-0000-4000-8000-000000000002', 'totp_v1',
    'totp_enroll_v1', 'totp_v1', 'hmac_sha256_v1',
    'iam-challenge-key-v1',
    pg_catalog.decode(pg_catalog.repeat('95', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('96', 32), 'hex'),
    'pending', 0, 1, '2026-07-16T00:20:00Z',
    '2026-07-16T00:25:00Z', '2026-07-17T00:25:00Z',
    '2026-07-16T00:20:00Z', '2026-07-16T00:20:00Z'
);

INSERT INTO iam_auth_challenges (
    id, user_id, session_id, purpose, proof_kind, hash_profile,
    hmac_key_id, challenge_hmac, context_sha256, step_up_purpose_sha256,
    action_policy_version, state, attempt_count, revision, issued_at,
    expires_at, purge_after, created_at, updated_at
) VALUES (
    '86000000-0000-4000-8000-000000000003',
    '51000000-0000-4000-8000-000000000003',
    '81000000-0000-4000-8000-000000000001',
    'reauth_v1', 'password_v1', 'hmac_sha256_v1',
    'iam-challenge-key-v1',
    pg_catalog.decode(pg_catalog.repeat('97', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('98', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('99', 32), 'hex'), 1,
    'pending', 0, 1, '2026-07-16T00:30:00Z',
    '2026-07-16T00:35:00Z', '2026-07-17T00:35:00Z',
    '2026-07-16T00:30:00Z', '2026-07-16T00:30:00Z'
);`
