//go:build integration

package postgres

const (
	iamOTPRoleID       = "85000000-0000-4000-8000-000000000001"
	iamOTPRoleScopeID  = "85000000-0000-4000-8000-000000000002"
	iamOTPIdentifierID = "86000000-0000-4000-8000-000000000001"
	iamOTPPortalID     = "86000000-0000-4000-8000-000000000002"
	iamOTPInvitationID = "87000000-0000-4000-8000-000000000001"
)

const insertIAMOTPRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    $1, 'tenant', $2, NULL, 'OTP Operator',
    pg_catalog.convert_to('OTP Operator', 'UTF8'), NULL, 'active',
    false, false, 'role', '41000000-0000-4000-8000-000000000003',
    '{}'::jsonb, pg_catalog.decode(pg_catalog.repeat('81', 32), 'hex'),
    1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
)`

const insertIAMInvitationFixture = `
INSERT INTO iam_invitations (
    id, identifier_kind, display_value, normalized_value, normalized_sha256,
    role_id, role_scope_kind, role_scope_id, organization_id, invited_by,
    accepted_user_id, state, revision, created_at, updated_at, expires_at,
    accepted_at, revoked_at, expired_at
) VALUES (
    $1, 'email_v1', $2, $3,
    pg_catalog.decode(pg_catalog.repeat($4, 32), 'hex'),
    $5, 'tenant', $6, NULL, $7, NULL, 'pending', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '7 days',
    NULL, NULL, NULL
)`

const insertIAMOTPPortalIdentifier = `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value, normalized_sha256,
    state, revision, verified_at, activated_at, tombstoned_at,
    cooldown_until, retired_at, created_at, updated_at
) VALUES (
    $1, $2, 'portal_v1', 'EMP-42', 'emp-42',
    pg_catalog.decode(pg_catalog.repeat('82', 32), 'hex'), 'active', 1,
    '2026-07-16T00:01:00Z', '2026-07-16T00:02:00Z', NULL, NULL, NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:02:00Z'
)`

const insertIAMOneTimeTokenFixture = `
INSERT INTO iam_one_time_tokens (
    id, purpose, user_id, identifier_id, identifier_kind, identifier_revision,
    invitation_id, invitation_expires_at, invitation_revision,
    portal_binding_id, portal_setup_revision,
    selector, hash_key_id, secret_hmac, binding_sha256, state, attempt_count,
    revision, created_at, updated_at, expires_at, consumed_at, revoked_at,
    expired_at, purge_after
) VALUES (
    $1, $2, $3, $4, $5,
    CASE WHEN $4::uuid IS NOT NULL THEN 1::bigint ELSE NULL END, $6,
    CASE WHEN $2 = 'invitation_accept_v1'
        THEN '2026-07-23T00:00:00Z'::timestamptz ELSE NULL END,
    CASE WHEN $2 = 'invitation_accept_v1' THEN 1::bigint ELSE NULL END,
    CASE WHEN $2 = 'portal_setup_v1'
        THEN '53000000-0000-4000-8000-000000000002'::uuid ELSE NULL END,
    CASE WHEN $2 = 'portal_setup_v1' THEN 1::bigint ELSE NULL END,
    pg_catalog.decode(pg_catalog.repeat($7, 16), 'hex'), 'otp-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat($8, 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat($9, 32), 'hex'),
    'pending', 0, 1, '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + $10::interval,
    NULL, NULL, NULL,
    '2026-07-16T00:00:00Z'::timestamptz + $10::interval + INTERVAL '24 hours'
)`
