//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAtomicRoutineACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		var routines, hardened, captured, runtime, adopter, public int
		err := database.QueryRowContext(ctx, `
WITH routines AS (
    SELECT routine.oid, routine.prosecdef, routine.proconfig, routine.proacl
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
    WHERE namespace.nspname = $1 AND routine.proname IN (
        'iam_refresh_rotate_v1', 'iam_one_time_consume_v1',
        'iam_invitation_accept_v1', 'iam_challenge_resolve_v1',
        'iam_recovery_consume_v1', 'iam_portal_setup_complete_v1'
    )
), public_acl AS (
    SELECT DISTINCT routine.oid
    FROM routines AS routine
    CROSS JOIN LATERAL pg_catalog.aclexplode(
        COALESCE(routine.proacl, pg_catalog.acldefault('f', 0))
    ) AS acl
    WHERE acl.grantee = 0 AND acl.privilege_type = 'EXECUTE'
)
SELECT pg_catalog.count(*),
       pg_catalog.count(*) FILTER (WHERE routine.prosecdef),
       pg_catalog.count(*) FILTER (
           WHERE routine.proconfig = ARRAY['search_path=' || $1 || ', pg_catalog']::text[]
       ),
       pg_catalog.count(*) FILTER (
           WHERE pg_catalog.has_function_privilege(
               'harpy_runtime', routine.oid, 'EXECUTE'
           )
       ),
       pg_catalog.count(*) FILTER (
           WHERE pg_catalog.has_function_privilege(
               'harpy_adopter', routine.oid, 'EXECUTE'
           )
       ),
       (SELECT pg_catalog.count(*) FROM public_acl)
FROM routines AS routine`, tenant.schema).Scan(
			&routines, &hardened, &captured, &runtime, &adopter, &public,
		)
		if err != nil {
			t.Fatalf("inspect %s IAM atomic routines: %v", tenant.schema, err)
		}
		if routines != 6 || hardened != 6 || captured != 6 ||
			runtime != 6 || adopter != 0 || public != 0 {
			t.Fatalf(
				"%s IAM routine ACL = %d/%d/%d/%d/%d/%d, want 6/6/6/6/0/0",
				tenant.schema, routines, hardened, captured, runtime, adopter, public,
			)
		}
		var runtimeCreate, publicCreate bool
		err = database.QueryRowContext(ctx, `
SELECT pg_catalog.has_schema_privilege('harpy_runtime', $1, 'CREATE'),
       pg_catalog.has_schema_privilege('public', $1, 'CREATE')`, tenant.schema).
			Scan(&runtimeCreate, &publicCreate)
		if err != nil || runtimeCreate || publicCreate {
			t.Fatalf("%s IAM schema CREATE runtime/public = %v/%v: %v",
				tenant.schema, runtimeCreate, publicCreate, err)
		}
	}
	assertIAMFoundationHasNoRuntimeDML(t, ctx, database)
}

func assertIAMAtomicRoutineScopeSecurity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	var outcome string
	err = transaction.QueryRowContext(ctx, `
SELECT t_iam_b.iam_one_time_consume_v1(
    't_iam_a', '88000000-0000-4000-8000-000000000001', 1,
    'password_reset_v1'
)`).Scan(&outcome)
	_ = transaction.Rollback()
	_ = connection.Close()
	assertIAMSQLState(t, err, "42501")
	assertIAMAtomicShadowSearchPath(t, ctx, database)

	connection, transaction, err = beginIAMAtomicRuntimeTx(ctx, database, "t_iam_b")
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
UPDATE iam_users SET revision = revision + 1
WHERE id = '51000000-0000-4000-8000-000000000001'`)
	_ = transaction.Rollback()
	_ = connection.Close()
	assertIAMSQLState(t, err, "42501")
}

func assertIAMAtomicShadowSearchPath(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE SCHEMA iam_attacker;
CREATE TABLE iam_attacker.iam_one_time_tokens (id uuid PRIMARY KEY, state text NOT NULL);
INSERT INTO iam_attacker.iam_one_time_tokens VALUES (
    '88000000-0000-4000-8000-000000000001', 'shadow_pending'
);
CREATE FUNCTION iam_attacker.iam_one_time_consume_v1(name,uuid,bigint,text)
RETURNS text LANGUAGE sql AS 'SELECT ''shadowed''::text';
GRANT USAGE ON SCHEMA iam_attacker TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_attacker.iam_one_time_consume_v1(
    name,uuid,bigint,text
) TO harpy_runtime;`)
	if err != nil {
		t.Fatalf("prepare IAM attacker schema: %v", err)
	}
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := transaction.ExecContext(
		ctx, "SELECT pg_catalog.set_config('search_path', 'iam_attacker,t_iam_a,pg_catalog', true)",
	); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	var outcome string
	err = transaction.QueryRowContext(ctx, `
SELECT t_iam_a.iam_one_time_consume_v1(
    't_iam_a', '88000000-0000-4000-8000-000000000001', 1,
    'password_reset_v1'
)`).Scan(&outcome)
	if err != nil || outcome != "consumed" {
		_ = transaction.Rollback()
		t.Fatalf("official IAM routine under attacker path = %q: %v", outcome, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var shadowState, tenantBState string
	if err := database.QueryRowContext(ctx,
		"SELECT state FROM iam_attacker.iam_one_time_tokens").Scan(&shadowState); err != nil {
		t.Fatal(err)
	}
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_b", `
SELECT state FROM iam_one_time_tokens
WHERE id = '88000000-0000-4000-8000-000000000001'`, &tenantBState)
	if shadowState != "shadow_pending" || tenantBState != "pending" {
		t.Fatalf("shadow/tenant-B state = %s/%s, want shadow_pending/pending",
			shadowState, tenantBState)
	}
}
