//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMSessions = `
INSERT INTO iam_sessions (
    id, user_id, access_mode_id, access_mode, portal_binding_id, state,
    device_kind, browser_family, os_family, ip_hmac_key_id, ip_hmac,
    revision, created_at, updated_at, last_seen_at, absolute_expires_at,
    revoked_at, expired_at, retain_until
) VALUES
    ('81000000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000003',
     '52000000-0000-4000-8000-000000000003', 'operator_console', NULL,
     'active', 'desktop', 'chromium', 'windows', 'ip-hmac-v1',
     pg_catalog.decode(pg_catalog.repeat('a1', 32), 'hex'), 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
     '2026-07-16T00:01:00Z',
     '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
     NULL, NULL,
     '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '120 days'),
    ('81000000-0000-4000-8000-000000000002',
     '51000000-0000-4000-8000-000000000002',
     '52000000-0000-4000-8000-000000000002', 'portal',
     '53000000-0000-4000-8000-000000000002',
     'active', 'mobile', 'safari', 'ios', 'ip-hmac-v1',
     pg_catalog.decode(pg_catalog.repeat('a2', 32), 'hex'), 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
     '2026-07-16T00:01:00Z',
     '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
     NULL, NULL,
     '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '120 days');`

const insertIAMRefreshFamily = `
INSERT INTO iam_refresh_families (
    id, user_id, session_id, state, rotation_sequence, revoke_reason,
    revision, created_at, updated_at, absolute_expires_at,
    revoked_at, compromised_at, expired_at, retain_until
) VALUES (
    '82000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003',
    '81000000-0000-4000-8000-000000000001', 'active', 0, NULL, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
    NULL, NULL, NULL,
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '91 days'
);`

const insertIAMRootRefreshToken = `
INSERT INTO iam_refresh_tokens (
    id, family_id, family_absolute_expires_at, parent_id, parent_sequence, selector,
    hash_key_id, secret_hmac, sequence, auth_time, acr,
    amr_password, amr_webauthn, amr_totp, amr_recovery,
    step_up_purpose_sha256, state, revision, issued_at, updated_at,
    idle_expires_at, consumed_at, revoked_at, expired_at, retain_until
) VALUES (
    '83000000-0000-4000-8000-000000000001',
    '82000000-0000-4000-8000-000000000001',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days', NULL, NULL,
    pg_catalog.decode(pg_catalog.repeat('b1', 16), 'hex'), 'refresh-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('c1', 32), 'hex'), 1,
    '2026-07-16T00:00:00Z', 'login', true, false, false, false,
    NULL, 'active', 1, '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '30 days',
    NULL, NULL, NULL,
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '91 days'
);`

const insertIAMSuccessorRefreshToken = `
INSERT INTO iam_refresh_tokens (
    id, family_id, family_absolute_expires_at, parent_id, parent_sequence, selector,
    hash_key_id, secret_hmac, sequence, auth_time, acr,
    amr_password, amr_webauthn, amr_totp, amr_recovery,
    step_up_purpose_sha256, state, revision, issued_at, updated_at,
    idle_expires_at, consumed_at, revoked_at, expired_at, retain_until
) VALUES (
    '83000000-0000-4000-8000-000000000002',
    '82000000-0000-4000-8000-000000000001',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
    '83000000-0000-4000-8000-000000000001',
    1,
    pg_catalog.decode(pg_catalog.repeat('b2', 16), 'hex'), 'refresh-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('c2', 32), 'hex'), 2,
    '2026-07-16T00:00:00Z', 'login', true, false, false, false,
    NULL, 'active', 1, '2026-07-16T01:00:00Z',
    '2026-07-16T01:00:00Z',
    '2026-07-16T01:00:00Z'::timestamptz + INTERVAL '30 days',
    NULL, NULL, NULL,
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '91 days'
);`

func assertIAMSessionRefreshConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMSessions)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMRefreshFamily)
		err := rejectIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_refresh_families (
    id, user_id, session_id, state, rotation_sequence, revision,
    created_at, updated_at, absolute_expires_at, retain_until
) VALUES (
    '82000000-0000-4000-8000-000000000099', $1,
    '81000000-0000-4000-8000-000000000002', 'active', 0, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '89 days',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '91 days'
)`, iamUserTwoID)
		assertIAMConstraint(t, err, "23514", "iam_refresh_families_time_check")

		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMRootRefreshToken)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_refresh_tokens
SET state = 'consumed', consumed_at = '2026-07-16T01:00:00Z',
    updated_at = '2026-07-16T01:00:00Z'
WHERE id = '83000000-0000-4000-8000-000000000001'`)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMSuccessorRefreshToken)
		err = rejectIAMTenant(t, ctx, database, tenant.schema, insertIAMGapRefreshToken)
		assertIAMConstraint(t, err, "23514", "iam_refresh_tokens_parent_check")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_refresh_tokens
SET state = 'active', consumed_at = NULL
WHERE id = '83000000-0000-4000-8000-000000000001'`)
		assertIAMConstraint(
			t, err, "23505", "iam_refresh_tokens_one_active_family_uidx",
		)
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_refresh_tokens
SET acr = 'step_up', amr_password = false, amr_recovery = true,
    step_up_purpose_sha256 = pg_catalog.decode(pg_catalog.repeat('d1', 32), 'hex')
WHERE id = '83000000-0000-4000-8000-000000000002'`)
		assertIAMConstraint(
			t, err, "23514", "iam_refresh_tokens_auth_context_check",
		)
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_refresh_tokens
SET acr = 'step_up', step_up_purpose_sha256 = NULL
WHERE id = '83000000-0000-4000-8000-000000000002'`)
		assertIAMConstraint(
			t, err, "23514", "iam_refresh_tokens_auth_context_check",
		)
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_sessions (
    id, user_id, access_mode_id, access_mode, portal_binding_id, state,
    device_kind, browser_family, os_family, ip_hmac_key_id, ip_hmac,
    revision, created_at, updated_at, last_seen_at, absolute_expires_at,
    retain_until
) VALUES (
    '81000000-0000-4000-8000-000000000099', $1, $2, 'portal',
    '53000000-0000-4000-8000-000000000002', 'active', 'desktop',
    'chromium', 'linux', 'ip-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('d2', 32), 'hex'), 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '90 days',
    '2026-07-16T00:00:00Z'::timestamptz + INTERVAL '120 days'
)`, iamUserThreeID, iamUserThreePortalModeID)
		assertIAMConstraint(t, err, "23503", "iam_sessions_portal_binding_fk")
	}
}
