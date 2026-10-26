//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMIdentifierNullTransition(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	err := rejectIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value,
    normalized_sha256, state, revision, verified_at, activated_at,
    tombstoned_at, cooldown_until, retired_at, created_at, updated_at
) VALUES (
    '61000000-0000-4000-8000-000000000005', $1, 'email_v1',
    'Missing@example.com', 'missing@example.com',
    pg_catalog.decode(pg_catalog.repeat('34', 32), 'hex'),
    'verified', 1, NULL, NULL, NULL, NULL, NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z'
)`, iamUserTwoID)
	assertIAMConstraint(
		t, err, "23514", "iam_user_identifiers_lifecycle_check",
	)
}
