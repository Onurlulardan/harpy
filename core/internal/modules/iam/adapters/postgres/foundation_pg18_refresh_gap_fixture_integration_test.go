//go:build integration

package postgres

const insertIAMGapRefreshToken = `
INSERT INTO iam_refresh_tokens (
    id, family_id, family_absolute_expires_at, parent_id, parent_sequence,
    selector, hash_key_id, secret_hmac, sequence, auth_time, acr,
    amr_password, amr_webauthn, amr_totp, amr_recovery,
    step_up_purpose_sha256, state, revision, issued_at, updated_at,
    idle_expires_at, consumed_at, retain_until
) VALUES (
    '83000000-0000-4000-8000-000000000099',
    '82000000-0000-4000-8000-000000000001',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
    '83000000-0000-4000-8000-000000000002', 2,
    pg_catalog.decode(pg_catalog.repeat('ef', 16), 'hex'), 'refresh-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('fe', 32), 'hex'), 4,
    '2026-07-16T00:00:00Z', 'login', true, false, false, false,
    NULL, 'consumed', 1, '2026-07-16T02:00:00Z',
    '2026-07-16T02:00:00Z',
    '2026-07-16T02:00:00Z'::timestamptz + INTERVAL '30 days',
    '2026-07-16T02:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '91 days'
)`
