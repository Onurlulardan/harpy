//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMRestQueryACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		var baseSelect, viewsSelect, routineExecute int
		err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FILTER (
           WHERE relation.relname IN ('iam_users','iam_roles','iam_organizations')
             AND pg_catalog.has_table_privilege(
                 'harpy_runtime', relation.oid, 'SELECT'
             )
       ),
       pg_catalog.count(*) FILTER (
           WHERE relation.relname IN (
                 'iam_rest_query_users_v1','iam_rest_query_roles_v1',
                 'iam_rest_query_organizations_v1','iam_rest_query_users_v2',
                 'iam_rest_query_roles_v2','iam_rest_query_organizations_v2',
                 'iam_rest_query_organizations_v3'
             )
             AND pg_catalog.has_table_privilege(
                 'harpy_runtime', relation.oid, 'SELECT'
             )
       ),
       pg_catalog.count(*) FILTER (
           WHERE relation.relname = 'iam_rest_query_actor_versions_v1'
             AND pg_catalog.has_function_privilege(
                 'harpy_runtime', relation.oid, 'EXECUTE'
             )
       )
FROM (
    SELECT class.oid, class.relname
    FROM pg_catalog.pg_class AS class
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = class.relnamespace
    WHERE namespace.nspname = $1
      AND class.relname IN (
          'iam_users','iam_roles','iam_organizations',
          'iam_rest_query_users_v1','iam_rest_query_roles_v1',
          'iam_rest_query_organizations_v1','iam_rest_query_users_v2',
          'iam_rest_query_roles_v2','iam_rest_query_organizations_v2',
          'iam_rest_query_organizations_v3'
      )
    UNION ALL
    SELECT routine.oid, routine.proname
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = routine.pronamespace
    WHERE namespace.nspname = $1
      AND routine.proname = 'iam_rest_query_actor_versions_v1'
) AS relation`, tenant.schema).Scan(&baseSelect, &viewsSelect, &routineExecute)
		if err != nil {
			t.Fatalf("inspect %s IAM REST Query ACL: %v", tenant.schema, err)
		}
		var publicSelect int
		err = database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(
        relation.relacl,
        pg_catalog.acldefault('r', relation.relowner)
    )
) AS privilege
WHERE namespace.nspname = $1
  AND relation.relname IN (
      'iam_rest_query_users_v1','iam_rest_query_roles_v1',
      'iam_rest_query_organizations_v1','iam_rest_query_users_v2',
      'iam_rest_query_roles_v2','iam_rest_query_organizations_v2',
      'iam_rest_query_organizations_v3'
  )
  AND privilege.grantee = 0
  AND privilege.privilege_type = 'SELECT'`, tenant.schema).Scan(&publicSelect)
		if err != nil {
			t.Fatalf("inspect %s IAM REST Query PUBLIC ACL: %v", tenant.schema, err)
		}
		if baseSelect != 0 || viewsSelect != 7 || publicSelect != 0 || routineExecute != 1 {
			t.Fatalf("%s IAM REST Query ACL = %d/%d/%d/%d, want 0/7/0/1",
				tenant.schema, baseSelect, viewsSelect, publicSelect, routineExecute)
		}
	}
}

func assertIAMRestQueryScopeSecurity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	var tokenVersion, authorizationVersion int64
	err = transaction.QueryRowContext(ctx, `
SELECT token_version, authorization_version
FROM iam_rest_query_actor_versions_v1(
    't_iam_b', '51000000-0000-4000-8000-000000000001'
)`).Scan(&tokenVersion, &authorizationVersion)
	_ = transaction.Rollback()
	_ = connection.Close()
	assertIAMSQLState(t, err, "42501")
}
