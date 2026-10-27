//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMDirectory = `
INSERT INTO platform.iam_login_directory (
    tenant_id, identifier_kind, lookup_key_id, identifier_hmac,
    user_id, source_identifier_id, source_revision, state,
    source_sha256, projected_at, tombstoned_at
) VALUES (
    $1, 'email_v1', 'directory-key-v1',
    pg_catalog.decode(pg_catalog.repeat('11', 32), 'hex'),
    $2, $3, 1, 'active',
    pg_catalog.decode(pg_catalog.repeat('22', 32), 'hex'),
    '2026-07-16T00:00:00Z', NULL
)`

func assertIAMLoginDirectoryConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	unknownErr := execIAMDirectory(
		ctx,
		database,
		"30000000-0000-4000-8000-000000000003",
		"31000000-0000-4000-8000-000000000003",
		"32000000-0000-4000-8000-000000000003",
	)
	assertIAMConstraint(t, unknownErr, "23503", "iam_login_directory_tenant_fk")

	if err := execIAMDirectory(
		ctx,
		database,
		iamFoundationTenants[0].id,
		"11000000-0000-4000-8000-000000000001",
		"12000000-0000-4000-8000-000000000001",
	); err != nil {
		t.Fatalf("insert valid IAM directory locator: %v", err)
	}
	duplicateErr := execIAMDirectory(
		ctx,
		database,
		iamFoundationTenants[0].id,
		"13000000-0000-4000-8000-000000000001",
		"14000000-0000-4000-8000-000000000001",
	)
	assertIAMConstraint(t, duplicateErr, "23505", "iam_login_directory_pkey")

	if err := execIAMDirectory(
		ctx,
		database,
		iamFoundationTenants[1].id,
		"21000000-0000-4000-8000-000000000002",
		"22000000-0000-4000-8000-000000000002",
	); err != nil {
		t.Fatalf("same directory digest in second tenant: %v", err)
	}
}

func execIAMDirectory(
	ctx context.Context,
	database *sql.DB,
	tenantID string,
	userID string,
	identifierID string,
) error {
	_, err := database.ExecContext(
		ctx,
		insertIAMDirectory,
		tenantID,
		userID,
		identifierID,
	)
	return err
}
