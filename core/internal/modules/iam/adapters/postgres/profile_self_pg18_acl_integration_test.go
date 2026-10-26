//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertIAMSelfProfileACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		var routines, definer, captured, runtime, adopter, public int
		err := database.QueryRowContext(ctx, `
WITH routines AS (
 SELECT routine.oid,routine.prosecdef,routine.proconfig
 FROM pg_catalog.pg_proc AS routine
 JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
 WHERE namespace.nspname=$1 AND routine.proname IN (
  'iam_self_profile_read_v1','iam_self_profile_update_v1')
)
SELECT pg_catalog.count(*),
 pg_catalog.count(*) FILTER (WHERE prosecdef),
 pg_catalog.count(*) FILTER (
  WHERE proconfig=ARRAY['search_path=' || $1 || ', pg_catalog']::text[]),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'harpy_runtime',oid,'EXECUTE')),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'harpy_adopter',oid,'EXECUTE')),
 pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
  'public',oid,'EXECUTE'))
FROM routines`, tenant.schema).Scan(
			&routines, &definer, &captured, &runtime, &adopter, &public,
		)
		if err != nil || routines != 2 || definer != 2 || captured != 2 ||
			runtime != 2 || adopter != 0 || public != 0 {
			t.Fatalf("%s self profile ACL=%d/%d/%d/%d/%d/%d: %v",
				tenant.schema, routines, definer, captured, runtime, adopter, public, err)
		}
	}
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, "t_iam_a",
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
UPDATE iam_users SET revision=revision+1
WHERE id='51000000-0000-4000-8000-000000000003'`)
	_ = transaction.Rollback()
	_ = connection.Close()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("runtime direct self profile DML error=%v", err)
	}
}
