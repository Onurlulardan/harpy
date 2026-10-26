//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func prepareIAMAtomicRuntimeRole(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
ALTER ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime
WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT USAGE ON SCHEMA t_iam_a TO harpy_runtime;
GRANT USAGE ON SCHEMA t_iam_b TO harpy_runtime;
REVOKE CREATE ON SCHEMA t_iam_a FROM PUBLIC;
REVOKE CREATE ON SCHEMA t_iam_b FROM PUBLIC;
REVOKE CREATE ON SCHEMA t_iam_a FROM harpy_runtime;
REVOKE CREATE ON SCHEMA t_iam_b FROM harpy_runtime;`)
	if err != nil {
		t.Fatalf("prepare IAM atomic runtime role: %v", err)
	}
}

func applyIAMAtomicRoutines(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamAtomicRoutineMigrationNames))
	for _, name := range iamAtomicRoutineMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
