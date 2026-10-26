//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsCatalog(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var platformTables, tenantTables int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables
WHERE table_schema='platform' AND table_name LIKE 'analytics\_%' ESCAPE '\'`).Scan(&platformTables); err != nil {
		t.Fatal(err)
	}
	if platformTables != 5 {
		t.Fatalf("Analytics platform tables = %d, want 5", platformTables)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables
WHERE table_schema='t_analytics_a' AND table_name LIKE 'analytics\_%' ESCAPE '\'`).Scan(&tenantTables); err != nil {
		t.Fatal(err)
	}
	if tenantTables < 10 {
		t.Fatalf("Analytics tenant tables = %d, want at least 10", tenantTables)
	}
	var wrongOwners, publicExecute int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM pg_proc AS routine
JOIN pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_roles AS owner ON owner.oid=routine.proowner
WHERE (namespace.nspname='platform' OR namespace.nspname='t_analytics_a')
  AND routine.proname LIKE 'analytics\_%' ESCAPE '\' AND owner.rolname<>'harpy_owner'`).Scan(&wrongOwners); err != nil {
		t.Fatal(err)
	}
	if wrongOwners != 0 {
		t.Fatalf("Analytics routines with wrong owner = %d", wrongOwners)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM pg_proc AS routine
JOIN pg_namespace AS namespace ON namespace.oid=routine.pronamespace
CROSS JOIN LATERAL aclexplode(COALESCE(routine.proacl,acldefault('f',routine.proowner))) AS privilege
WHERE (namespace.nspname='platform' OR namespace.nspname='t_analytics_a')
  AND routine.proname LIKE 'analytics\_%' ESCAPE '\'
  AND privilege.grantee=0 AND privilege.privilege_type='EXECUTE'`).Scan(&publicExecute); err != nil {
		t.Fatal(err)
	}
	if publicExecute != 0 {
		t.Fatalf("Analytics PUBLIC routine grants = %d", publicExecute)
	}
}
