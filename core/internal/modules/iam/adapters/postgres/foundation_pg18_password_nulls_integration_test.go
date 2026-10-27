//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMEmptyCredentialShape = `
INSERT INTO iam_password_credentials (
    user_id, credential_version, state, algorithm, salt, tag,
    memory_kib, iterations, parallelism, legacy_bcrypt_hash,
    upgrade_required, revision, created_at, updated_at,
    last_used_at, invalidated_at
) VALUES (
    $1, 1, 'active', $2, NULL, NULL, NULL, NULL, NULL, $3, $4, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, NULL
)`

func assertIAMPasswordNullShapes(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	argonErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMEmptyCredentialShape,
		iamUserThreeID, "argon2id_v1", nil, false,
	)
	assertIAMConstraint(
		t, argonErr, "23514", "iam_password_credentials_shape_check",
	)
	bcryptErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMEmptyCredentialShape,
		iamUserThreeID, "bcrypt_v1_adopted", nil, true,
	)
	assertIAMConstraint(
		t, bcryptErr, "23514", "iam_password_credentials_shape_check",
	)
	badCostErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMEmptyCredentialShape,
		iamUserThreeID, "bcrypt_v1_adopted",
		"$2b$03$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1", true,
	)
	assertIAMConstraint(
		t, badCostErr, "23514", "iam_password_credentials_shape_check",
	)
}
