//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const seedIAMExpiredPermissionRequest = `
INSERT INTO iam_permission_request_groups (
    id, requester_user_id, request_mode, creator_authority_permission_id,
    reviewer_authority_permission_id, step_up_purpose_sha256, state,
    request_count, reviewer_user_id, review_task_id, revision, expires_at,
    terminal_at, created_at, updated_at
) VALUES (
    'ae000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002', 'SELF_SERVICE',
    'a0030000-0000-4000-8000-000000000001',
    'a0030000-0000-4000-8000-000000000002',
    pg_catalog.decode(pg_catalog.repeat('a5',32),'hex'), 'PENDING',
    1, NULL, NULL, 1, pg_catalog.transaction_timestamp() - INTERVAL '1 hour',
    NULL, pg_catalog.transaction_timestamp() - INTERVAL '1 day',
    pg_catalog.transaction_timestamp() - INTERVAL '1 day'
);
INSERT INTO iam_permission_requests (
    id, group_id, ordinal, target_kind, target_user_id, target_role_id,
    target_organization_id, permission_id, permission_match_kind,
    permission_is_protected, planned_binding_id, granted_binding_id,
    state, revision, terminal_at, created_at, updated_at
) VALUES (
    'ae100000-0000-4000-8000-000000000001',
    'ae000000-0000-4000-8000-000000000001', 1, 'USER',
    '51000000-0000-4000-8000-000000000002', NULL, NULL,
    '93000000-0000-4000-8000-000000000001', 'exact', false,
    'ae200000-0000-4000-8000-000000000001', NULL,
    'PENDING', 1, NULL, pg_catalog.transaction_timestamp() - INTERVAL '1 day',
    pg_catalog.transaction_timestamp() - INTERVAL '1 day'
);`

const seedIAMExpiredInvitation = `
INSERT INTO iam_invitations (
    id, identifier_kind, display_value, normalized_value, normalized_sha256,
    role_id, role_scope_kind, role_scope_id, organization_id, invited_by,
    accepted_user_id, accepted_identifier_id, accepted_identifier_kind,
    state, revision, created_at, updated_at, expires_at,
    accepted_at, revoked_at, expired_at
) VALUES (
    'be000000-0000-4000-8000-000000000001', 'email_v1',
    'expired@example.com', 'expired@example.com',
    pg_catalog.decode(pg_catalog.repeat('b5',32),'hex'),
    'b0040000-0000-4000-8000-000000000001', 'tenant', $1, NULL,
    '51000000-0000-4000-8000-000000000001', NULL, NULL, NULL,
    'pending', 1, pg_catalog.transaction_timestamp() - INTERVAL '8 days',
    pg_catalog.transaction_timestamp() - INTERVAL '8 days',
    pg_catalog.transaction_timestamp() - INTERVAL '1 day', NULL, NULL, NULL
);`

const seedIAMExpiredInvitationToken = `
INSERT INTO iam_one_time_tokens (
    id, purpose, user_id, identifier_id, identifier_kind,
    identifier_revision, invitation_id, invitation_expires_at,
    invitation_revision, portal_binding_id, portal_setup_revision,
    selector, hash_key_id, secret_hmac, binding_sha256, state,
    attempt_count, revision, created_at, updated_at, expires_at,
    consumed_at, revoked_at, expired_at, purge_after
) VALUES (
    'be100000-0000-4000-8000-000000000001', 'invitation_accept_v1',
    NULL, NULL, NULL, NULL, 'be000000-0000-4000-8000-000000000001',
    pg_catalog.transaction_timestamp() - INTERVAL '1 day', 1, NULL, NULL,
    pg_catalog.decode('be000000000000000000000000000001','hex'),
    'one-time-1', pg_catalog.decode(pg_catalog.repeat('b6',32),'hex'),
    pg_catalog.decode(pg_catalog.repeat('b7',32),'hex'), 'pending', 0, 1,
    pg_catalog.transaction_timestamp() - INTERVAL '8 days',
    pg_catalog.transaction_timestamp() - INTERVAL '8 days',
    pg_catalog.transaction_timestamp() - INTERVAL '1 day', NULL, NULL, NULL,
    pg_catalog.transaction_timestamp()
);`

func seedIAMPermissionInvitationExpiryFixtures(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer connection.Close()
	if _, err := transaction.ExecContext(ctx, seedIAMExpiredPermissionRequest); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("seed expired permission request in %s: %v", tenant.schema, err)
	}
	if _, err := transaction.ExecContext(ctx, seedIAMExpiredInvitation, tenant.id); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("seed expired invitation in %s: %v", tenant.schema, err)
	}
	if _, err := transaction.ExecContext(ctx, seedIAMExpiredInvitationToken); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("seed expired invitation token in %s: %v", tenant.schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit expiry fixtures in %s: %v", tenant.schema, err)
	}
}
