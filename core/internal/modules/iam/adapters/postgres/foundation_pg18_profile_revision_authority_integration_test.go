//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMProfileRevisionAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		assertIAMProfileRoutineACL(t, ctx, database, tenant.schema)
	}
	assertIAMFoundationHasNoRuntimeDML(t, ctx, database)
	assertIAMProfileRuntimeCannotReadTables(t, ctx, database)
	assertIAMProfileAdopterCannotExecute(t, ctx, database)
	assertIAMProfileScopeMismatchRejected(t, ctx, database)
}

func assertIAMProfileRoutineACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	var routines, definer, captured, runtime, adopter, public int
	err := database.QueryRowContext(ctx, `
WITH routines AS (
    SELECT routine.oid, routine.prosecdef, routine.proconfig, routine.proacl
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = routine.pronamespace
    WHERE namespace.nspname = $1 AND routine.proname IN (
        'iam_profile_historical_read_v1',
        'iam_profile_active_catalog_read_v1',
        'iam_profile_document_read_v1'
    )
), public_acl AS (
    SELECT DISTINCT routine.oid FROM routines AS routine
    CROSS JOIN LATERAL pg_catalog.aclexplode(
        COALESCE(routine.proacl, pg_catalog.acldefault('f', 0))
    ) AS acl
    WHERE acl.grantee = 0 AND acl.privilege_type = 'EXECUTE'
)
SELECT pg_catalog.count(*),
       pg_catalog.count(*) FILTER (WHERE routine.prosecdef),
       pg_catalog.count(*) FILTER (WHERE routine.proconfig =
           ARRAY['search_path=' || $1 || ', pg_catalog']::text[]),
       pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
           'harpy_runtime', routine.oid, 'EXECUTE')),
       pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
           'harpy_adopter', routine.oid, 'EXECUTE')),
       (SELECT pg_catalog.count(*) FROM public_acl)
FROM routines AS routine`, schema).Scan(
		&routines, &definer, &captured, &runtime, &adopter, &public)
	if err != nil || routines != 3 || definer != 3 || captured != 3 ||
		runtime != 3 || adopter != 0 || public != 0 {
		t.Fatalf("%s profile routine ACL=%d/%d/%d/%d/%d/%d: %v",
			schema, routines, definer, captured, runtime, adopter, public, err)
	}
}

func assertIAMProfileRuntimeCannotReadTables(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	var count int
	err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.count(*) FROM iam_profile_schema_versions").Scan(&count)
	_ = transaction.Rollback()
	_ = connection.Close()
	assertIAMSQLState(t, err, "42501")
}

func assertIAMProfileAdopterCannotExecute(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, "t_iam_a")
	defer connection.Close()
	_, err := transaction.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION harpy_adopter")
	if err == nil {
		var payload string
		err = transaction.QueryRowContext(ctx, `
SELECT t_iam_a.iam_profile_historical_read_v1(
    pg_catalog.current_schema()::name, $1
)::text`, iamUserProfileSchemaID).Scan(&payload)
	}
	_ = transaction.Rollback()
	assertIAMSQLState(t, err, "42501")
}

func assertIAMProfileScopeMismatchRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	var payload string
	err = transaction.QueryRowContext(ctx, `
SELECT t_iam_b.iam_profile_historical_read_v1(
    't_iam_a'::name, $1
)::text`, iamUserProfileSchemaID).Scan(&payload)
	_ = transaction.Rollback()
	_ = connection.Close()
	assertIAMSQLState(t, err, "42501")
}
