//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMActiveIdentifier = `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value,
    normalized_sha256, state, revision, verified_at, activated_at,
    tombstoned_at, cooldown_until, retired_at, created_at, updated_at
) VALUES (
    $1, $2, 'email_v1', $3, $4,
    pg_catalog.decode(pg_catalog.repeat('31', 32), 'hex'),
    'active', 1, '2026-07-16T00:01:00Z', '2026-07-16T00:02:00Z',
    NULL, NULL, NULL, '2026-07-16T00:00:00Z', '2026-07-16T00:02:00Z'
)`

func assertIAMIdentifierConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const sharedIdentifierID = "61000000-0000-4000-8000-000000000001"
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMActiveIdentifier,
			sharedIdentifierID, iamUserOneID,
			"Shared.User@example.com", "shared.user@example.com",
		)
	}

	reservedErr := rejectIAMTenant(t, ctx, database, "t_iam_a", `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value,
    normalized_sha256, state, revision, verified_at, activated_at,
    tombstoned_at, cooldown_until, retired_at, created_at, updated_at
) VALUES (
    '61000000-0000-4000-8000-000000000002', $1, 'email_v1',
    'Shared.User@example.com', 'shared.user@example.com',
    pg_catalog.decode(pg_catalog.repeat('32', 32), 'hex'),
    'pending', 1, NULL, NULL, NULL, NULL, NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
)`, iamUserTwoID)
	assertIAMConstraint(
		t, reservedErr, "23505", "iam_user_identifiers_reserved_value_uidx",
	)

	activeErr := rejectIAMTenant(
		t, ctx, database, "t_iam_a", insertIAMActiveIdentifier,
		"61000000-0000-4000-8000-000000000003", iamUserOneID,
		"Second@example.com", "second@example.com",
	)
	assertIAMConstraint(
		t, activeErr, "23505", "iam_user_identifiers_one_active_uidx",
	)

	cooldownErr := rejectIAMTenant(t, ctx, database, "t_iam_a", `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value,
    normalized_sha256, state, revision, verified_at, activated_at,
    tombstoned_at, cooldown_until, retired_at, created_at, updated_at
) VALUES (
    '61000000-0000-4000-8000-000000000004', $1, 'email_v1',
    'Cooldown@example.com', 'cooldown@example.com',
    pg_catalog.decode(pg_catalog.repeat('33', 32), 'hex'),
    'tombstoned', 1, '2026-07-16T00:01:00Z', '2026-07-16T00:02:00Z',
    '2026-07-16T00:03:00Z', '2026-07-16T00:03:00Z', NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:03:00Z'
)`, iamUserTwoID)
	assertIAMConstraint(
		t, cooldownErr, "23514", "iam_user_identifiers_lifecycle_check",
	)
	assertIAMIdentifierNullTransition(t, ctx, database, "t_iam_a")
}
