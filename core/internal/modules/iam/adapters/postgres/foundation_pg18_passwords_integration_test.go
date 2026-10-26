//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMArgon2Credential = `
INSERT INTO iam_password_credentials (
    user_id, credential_version, state, algorithm, salt, tag,
    memory_kib, iterations, parallelism, legacy_bcrypt_hash,
    upgrade_required, revision, created_at, updated_at,
    last_used_at, invalidated_at
) VALUES (
    $1, 1, 'active', 'argon2id_v1',
    pg_catalog.decode(pg_catalog.repeat('51', 16), 'hex'),
    pg_catalog.decode(pg_catalog.repeat('52', 32), 'hex'),
    65536, 3, 4, $2, $3, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, NULL
)`

const insertIAMBcryptCredential = `
INSERT INTO iam_password_credentials (
    user_id, credential_version, state, algorithm, salt, tag,
    memory_kib, iterations, parallelism, legacy_bcrypt_hash,
    upgrade_required, revision, created_at, updated_at,
    last_used_at, invalidated_at
) VALUES (
    $1, 1, 'active', 'bcrypt_v1_adopted',
    NULL, NULL, NULL, NULL, NULL,
    '$2b$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1',
    true, 1, '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, NULL
)`

func assertIAMPasswordCredentialConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const schema = "t_iam_b"
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMArgon2Credential,
		iamUserOneID, nil, false,
	)
	mustExecIAMTenant(
		t, ctx, database, schema, insertIAMBcryptCredential,
		iamUserTwoID,
	)

	hybridErr := rejectIAMTenant(
		t, ctx, database, schema, insertIAMArgon2Credential,
		iamUserThreeID,
		"$2b$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1",
		false,
	)
	assertIAMConstraint(
		t, hybridErr, "23514", "iam_password_credentials_shape_check",
	)
	assertIAMPasswordNullShapes(t, ctx, database, schema)

	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	var argon2Rows, bcryptRows int
	if err := transaction.QueryRowContext(ctx, `
SELECT
    pg_catalog.count(*) FILTER (
        WHERE algorithm = 'argon2id_v1' AND
              pg_catalog.octet_length(salt) = 16 AND
              pg_catalog.octet_length(tag) = 32 AND
              memory_kib = 65536 AND iterations = 3 AND parallelism = 4 AND
              legacy_bcrypt_hash IS NULL AND NOT upgrade_required
    ),
    pg_catalog.count(*) FILTER (
        WHERE algorithm = 'bcrypt_v1_adopted' AND
              salt IS NULL AND tag IS NULL AND memory_kib IS NULL AND
              iterations IS NULL AND parallelism IS NULL AND upgrade_required
    )
FROM iam_password_credentials`).Scan(&argon2Rows, &bcryptRows); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("read IAM credential shapes: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM credential assertion: %v", err)
	}
	if argon2Rows != 1 || bcryptRows != 1 {
		t.Fatalf("IAM credential shapes = argon2:%d bcrypt:%d", argon2Rows, bcryptRows)
	}
}
