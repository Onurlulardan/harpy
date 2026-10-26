//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMAtomicInvitationFixture = `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value, normalized_sha256,
    state, revision, verified_at, activated_at, created_at, updated_at
) VALUES (
    '96000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003', 'email_v1',
    'Routine.Invitee@example.com', 'routine.invitee@example.com',
    pg_catalog.decode(pg_catalog.repeat('c1', 32), 'hex'), 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
);
INSERT INTO iam_invitations (
    id, identifier_kind, display_value, normalized_value, normalized_sha256,
    role_id, role_scope_kind, role_scope_id, invited_by, state, revision,
    created_at, updated_at, expires_at
) VALUES (
    '97000000-0000-4000-8000-000000000001', 'email_v1',
    'Routine.Invitee@example.com', 'routine.invitee@example.com',
    pg_catalog.decode(pg_catalog.repeat('c1', 32), 'hex'),
    '85000000-0000-4000-8000-000000000001', 'tenant',
    '85000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000001', 'pending', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp() + INTERVAL '7 days'
);
INSERT INTO iam_one_time_tokens (
    id, purpose, invitation_id, invitation_expires_at, invitation_revision,
    selector, hash_key_id, secret_hmac, binding_sha256, state, attempt_count,
    revision, created_at, updated_at, expires_at, purge_after
)
SELECT '98000000-0000-4000-8000-000000000001', 'invitation_accept_v1',
       invitation.id, invitation.expires_at, invitation.revision,
       pg_catalog.decode(pg_catalog.repeat('c2', 16), 'hex'), 'otp-hmac-v1',
       pg_catalog.decode(pg_catalog.repeat('c3', 32), 'hex'),
       pg_catalog.decode(pg_catalog.repeat('c4', 32), 'hex'),
       'pending', 0, 1, invitation.created_at, invitation.created_at,
       invitation.expires_at, invitation.expires_at + INTERVAL '24 hours'
FROM iam_invitations AS invitation
WHERE invitation.id = '97000000-0000-4000-8000-000000000001';`

const insertIAMAtomicOneTimeFixtures = `
UPDATE iam_one_time_tokens
SET created_at = pg_catalog.transaction_timestamp(),
    updated_at = pg_catalog.transaction_timestamp(),
    expires_at = pg_catalog.transaction_timestamp() + INTERVAL '1 hour',
    purge_after = pg_catalog.transaction_timestamp() + INTERVAL '25 hours'
WHERE id = '88000000-0000-4000-8000-000000000001';
UPDATE iam_one_time_tokens
SET created_at = pg_catalog.transaction_timestamp(),
    updated_at = pg_catalog.transaction_timestamp(),
    expires_at = pg_catalog.transaction_timestamp() + INTERVAL '24 hours',
    purge_after = pg_catalog.transaction_timestamp() + INTERVAL '48 hours'
WHERE id = '88000000-0000-4000-8000-000000000002';`

const insertIAMAtomicPortalFixture = `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value, normalized_sha256,
    state, revision, verified_at, activated_at, created_at, updated_at
) VALUES (
    '96000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000001', 'portal_v1',
    'ROUTINE-PORTAL', 'routine-portal',
    pg_catalog.decode(pg_catalog.repeat('cb', 32), 'hex'), 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
);
INSERT INTO iam_one_time_tokens (
    id, purpose, user_id, identifier_id, identifier_kind, identifier_revision,
    portal_binding_id, portal_setup_revision, selector, hash_key_id,
    secret_hmac, binding_sha256, state, attempt_count, revision,
    created_at, updated_at, expires_at, purge_after
) VALUES (
    '98000000-0000-4000-8000-000000000004', 'portal_setup_v1',
    '51000000-0000-4000-8000-000000000001',
    '96000000-0000-4000-8000-000000000002', 'portal_v1', 1,
    '53000000-0000-4000-8000-000000000003', 1,
    pg_catalog.decode(pg_catalog.repeat('cc', 16), 'hex'), 'otp-hmac-v1',
    pg_catalog.decode(pg_catalog.repeat('cd', 32), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('ce', 32), 'hex'), 'pending', 0, 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp() + INTERVAL '1 hour',
    pg_catalog.transaction_timestamp() + INTERVAL '25 hours'
);`

func seedIAMAtomicRoutineFixtures(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAtomicInvitationFixture)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAtomicOneTimeFixtures)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAtomicPortalFixture)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAtomicChallengeFixture)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMAtomicRefreshFixture)
	}
}
