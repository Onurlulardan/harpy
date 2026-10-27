//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMLifecycleGuardTerminality(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	assertIAMSessionFamilyTerminality(t, ctx, database, schema)
	assertIAMRefreshRecoveryTerminality(t, ctx, database, schema)
	assertIAMChallengeTokenTerminality(t, ctx, database, schema)
}

func assertIAMSessionFamilyTerminality(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_sessions
SET state = 'revoked', revoked_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = 2
WHERE id = '81000000-0000-4000-8000-000000000001'`)
	err := rejectIAMTenantReplica(t, ctx, database, schema, `
UPDATE iam_sessions
SET state = 'active', revoked_at = NULL,
    updated_at = '2026-07-16T02:01:00Z', revision = 3
WHERE id = '81000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_sessions_state_transition_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_families
SET state = 'revoked', revoke_reason = 'manual',
    revoked_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = 3
WHERE id = '82000000-0000-4000-8000-000000000001'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_families
SET state = 'active', revoke_reason = NULL, revoked_at = NULL,
    updated_at = '2026-07-16T02:01:00Z', revision = 4
WHERE id = '82000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_refresh_families_state_transition_guard")
}

func assertIAMRefreshRecoveryTerminality(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_tokens
SET state = 'consumed', consumed_at = '2026-07-16T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = 2
WHERE id = '83000000-0000-4000-8000-000000000002'`)
	err := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_refresh_tokens
SET state = 'active', consumed_at = NULL,
    updated_at = '2026-07-16T02:01:00Z', revision = 3
WHERE id = '83000000-0000-4000-8000-000000000002'`)
	assertIAMConstraint(t, err, "23514", "iam_refresh_tokens_state_transition_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_recovery_codes
SET state = 'consumed', consumed_at = '2026-07-16T02:00:00Z',
    purge_after = '2026-08-15T02:00:00Z',
    updated_at = '2026-07-16T02:00:00Z', revision = 2
WHERE id = '85000000-0000-4000-8000-000000000001'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_recovery_codes
SET state = 'active', consumed_at = NULL, purge_after = NULL,
    updated_at = '2026-07-16T02:01:00Z', revision = 3
WHERE id = '85000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_recovery_codes_state_transition_guard")
}

func assertIAMChallengeTokenTerminality(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_auth_challenges
SET state = 'consumed', terminal_at = '2026-07-16T00:14:00Z',
    updated_at = '2026-07-16T00:14:00Z', revision = 2
WHERE id = '86000000-0000-4000-8000-000000000001'`)
	err := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_auth_challenges
SET state = 'pending', terminal_at = NULL,
    updated_at = '2026-07-16T00:14:30Z', revision = 3
WHERE id = '86000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_auth_challenges_state_transition_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_one_time_tokens
SET state = 'consumed', consumed_at = '2026-07-16T00:30:00Z',
    updated_at = '2026-07-16T00:30:00Z', revision = 2
WHERE id = '88000000-0000-4000-8000-000000000001'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_one_time_tokens
SET state = 'pending', consumed_at = NULL,
    updated_at = '2026-07-16T00:31:00Z', revision = 3
WHERE id = '88000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_state_transition_guard")

	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_invitations
SET state = 'revoked', revoked_at = '2026-07-16T01:00:00Z',
    updated_at = '2026-07-16T01:00:00Z', revision = 2
WHERE id = '87000000-0000-4000-8000-000000000001'`)
	err = rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_invitations
SET state = 'pending', revoked_at = NULL,
    updated_at = '2026-07-16T01:01:00Z', revision = 3
WHERE id = '87000000-0000-4000-8000-000000000001'`)
	assertIAMConstraint(t, err, "23514", "iam_invitations_state_transition_guard")
}
