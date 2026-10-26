//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertIAMRBACHTTPACLAndIsolation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		var routines, definer, captured, runtime, adopter, public int
		err := database.QueryRowContext(ctx, `
WITH routines AS (
 SELECT routine.oid, routine.prosecdef, routine.proconfig
 FROM pg_catalog.pg_proc AS routine
 JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
 WHERE namespace.nspname=$1 AND routine.proname IN (
  'iam_rbac_assignment_create_v1','iam_rbac_assignment_update_v1',
  'iam_rbac_assignment_retire_v1','iam_rbac_assignment_read_v1',
  'iam_rbac_assignment_list_v1','iam_rbac_effective_permissions_v1'
 )
)
SELECT pg_catalog.count(*),
 pg_catalog.count(*) FILTER (WHERE prosecdef),
 pg_catalog.count(*) FILTER (
  WHERE proconfig=ARRAY['search_path=' || $1 || ', pg_catalog']::text[]),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'harpy_runtime', oid, 'EXECUTE')),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'harpy_adopter', oid, 'EXECUTE')),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'public', oid, 'EXECUTE'))
FROM routines`, tenant.schema).Scan(
			&routines, &definer, &captured, &runtime, &adopter, &public,
		)
		if err != nil || routines != 6 || definer != 6 || captured != 6 ||
			runtime != 6 || adopter != 0 || public != 0 {
			t.Fatalf("%s assignment ACL=%d/%d/%d/%d/%d/%d: %v",
				tenant.schema, routines, definer, captured, runtime, adopter, public, err)
		}
	}
	assertIAMRBACHTTPTenantIdentity(t, ctx, database)
	assertIAMRBACHTTPRuntimeCannotWriteTables(t, ctx, database)
	assertIAMRBACHTTPForeignScopeDenied(t, ctx, database)
}

func assertIAMRBACHTTPTenantIdentity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, schema := range []string{"t_iam_a", "t_iam_b"} {
		var count int
		mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT pg_catalog.count(*) FROM iam_rbac_assignment_groups
WHERE id='b2000000-0000-4000-8000-000000000001'`, &count)
		if count != 1 {
			t.Fatalf("%s assignment identity count=%d", schema, count)
		}
	}
}

func assertIAMRBACHTTPRuntimeCannotWriteTables(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
INSERT INTO iam_rbac_assignment_groups (id) VALUES (
 'b9000000-0000-4000-8000-000000000001')`)
	_ = transaction.Rollback()
	_ = connection.Close()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("runtime direct assignment DML error=%v", err)
	}
}

func assertIAMRBACHTTPForeignScopeDenied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := runIAMAtomicText(ctx, database, "t_iam_a", `
SELECT pg_catalog.count(*)::text
FROM t_iam_b.iam_rbac_assignment_list_v1(
 't_iam_a'::name, 'all', NULL, 10,
 '51000000-0000-4000-8000-000000000001',
 '93000000-0000-4000-8000-000000000001')`)
	if err == nil {
		t.Fatal("assignment read accepted a foreign tenant scope")
	}
}
