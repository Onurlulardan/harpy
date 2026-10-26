//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

var iamPlatformRoutineSecurity = map[string]bool{
	"platform.iam_assert_current_tenant_authority_v1(uuid,bigint,bigint)": true,
	"platform.iam_refresh_locator_guard_v1()":                             false,
	"platform.iam_refresh_locator_create_v1(bytea,uuid,uuid,uuid,bigint,bigint,bigint,bytea,text,timestamp with time zone,timestamp with time zone,timestamp with time zone,timestamp with time zone)": true,
	"platform.iam_refresh_locator_advance_v1(bytea,uuid,bigint,bigint,bigint,bigint,bytea,text,timestamp with time zone,timestamp with time zone)":                                                     true,
	"platform.iam_auth_throttle_guard_v1()":                                                                                   false,
	"platform.iam_auth_throttle_policy_v1(text,text,text,integer,integer,integer,integer)":                                    false,
	"platform.iam_tenant_throttle_record_attempt_v1(uuid,text,text,text,bytea,bigint,bigint,integer,integer,integer,integer)": true,
	"platform.iam_environment_throttle_record_attempt_v1(text,text,text,bytea,integer,integer,integer,integer)":               true,
	"platform.iam_resolve_tenant_target_v1(uuid)":                                                                             true,
	"platform.iam_resolve_login_target_v1(uuid,text,text,bytea)":                                                              true,
	"platform.iam_resolve_refresh_target_v1(bytea)":                                                                           true,
	"platform.iam_resolve_login_target_v2(uuid,text,text,bytea)":                                                              true,
	"platform.iam_resolve_refresh_target_v2(bytea)":                                                                           true,
}

var iamRuntimePlatformRoutines = map[string]bool{
	"platform.iam_refresh_locator_create_v1(bytea,uuid,uuid,uuid,bigint,bigint,bigint,bytea,text,timestamp with time zone,timestamp with time zone,timestamp with time zone,timestamp with time zone)": true,
	"platform.iam_refresh_locator_advance_v1(bytea,uuid,bigint,bigint,bigint,bigint,bytea,text,timestamp with time zone,timestamp with time zone)":                                                     true,
	"platform.iam_tenant_throttle_record_attempt_v1(uuid,text,text,text,bytea,bigint,bigint,integer,integer,integer,integer)":                                                                          true,
	"platform.iam_environment_throttle_record_attempt_v1(text,text,text,bytea,integer,integer,integer,integer)":                                                                                        true,
	"platform.iam_resolve_tenant_target_v1(uuid)":                true,
	"platform.iam_resolve_login_target_v2(uuid,text,text,bytea)": true,
	"platform.iam_resolve_refresh_target_v2(bytea)":              true,
}

func assertIAMPlatformGuardCatalog(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	assertIAMPlatformGuardSessions(t, ctx, database)
	assertIAMPlatformRoleContract(t, ctx, database)
	rows, err := database.admin.QueryContext(ctx, `
SELECT procedure.oid::regprocedure::text, procedure.prosecdef, owner.rolname,
       owner.rolsuper, owner.rolcanlogin, procedure.proconfig::text,
       pg_catalog.has_function_privilege(
           'harpy_runtime', procedure.oid, 'EXECUTE'),
       pg_catalog.has_function_privilege(
           'harpy_adopter', procedure.oid, 'EXECUTE')
FROM pg_catalog.pg_proc AS procedure
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=procedure.proowner
WHERE namespace.nspname='platform' AND procedure.proname LIKE 'iam%'
ORDER BY procedure.proname`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := make(map[string]bool)
	for rows.Next() {
		var name, owner, config string
		var securityDefiner, superuser, login, runtimeExecute, adopterExecute bool
		if err := rows.Scan(&name, &securityDefiner, &owner, &superuser, &login,
			&config, &runtimeExecute, &adopterExecute); err != nil {
			t.Fatal(err)
		}
		expectedDefiner, expected := iamPlatformRoutineSecurity[name]
		if !expected || securityDefiner != expectedDefiner ||
			owner != "harpy_owner" || superuser || login ||
			config != "{search_path=pg_catalog}" || adopterExecute ||
			runtimeExecute != iamRuntimePlatformRoutines[name] {
			t.Fatalf("unexpected routine ACL/security for %s", name)
		}
		seen[name] = true
	}
	if err := rows.Err(); err != nil || len(seen) != len(iamPlatformRoutineSecurity) {
		t.Fatalf("IAM platform routine catalog count=%d err=%v", len(seen), err)
	}
	assertIAMPlatformTriggerCatalog(t, ctx, database)
	assertIAMPlatformTableACL(t, ctx, database)
}

func assertIAMPlatformGuardSessions(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	assertIAMPlatformSession(t, ctx, database.runtime,
		"harpy_v2_runtime", "harpy_runtime")
	assertIAMPlatformSession(t, ctx, database.migrator,
		"harpy_migrator", "harpy_owner")
}

func assertIAMPlatformSession(
	t *testing.T, ctx context.Context, database *sql.DB,
	wantSession, wantCurrent string,
) {
	t.Helper()
	var session, current string
	if err := database.QueryRowContext(ctx,
		"SELECT session_user,current_user").Scan(&session, &current); err != nil {
		t.Fatal(err)
	}
	if session != wantSession || current != wantCurrent {
		t.Fatalf("session/current=%s/%s, want %s/%s",
			session, current, wantSession, wantCurrent)
	}
}

func assertIAMPlatformRoleContract(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var rolesValid, membershipsValid, schemaValid bool
	err := database.admin.QueryRowContext(ctx, `
SELECT count(*)=5 AND bool_and(
 NOT rolsuper AND NOT rolinherit AND NOT rolbypassrls AND NOT rolcreatedb AND
 NOT rolcreaterole AND NOT rolreplication AND
 rolcanlogin=(rolname IN ('harpy_migrator','harpy_v2_runtime','harpy_adopter')))
FROM pg_catalog.pg_roles WHERE rolname IN (
 'harpy_owner','harpy_migrator','harpy_runtime',
 'harpy_v2_runtime','harpy_adopter')`).Scan(&rolesValid)
	if err != nil {
		t.Fatal(err)
	}
	err = database.admin.QueryRowContext(ctx, `
SELECT count(*)=2 AND bool_and(
 NOT membership.admin_option AND NOT membership.inherit_option AND membership.set_option AND
 ((granted.rolname='harpy_owner' AND member.rolname='harpy_migrator') OR
  (granted.rolname='harpy_runtime' AND member.rolname='harpy_v2_runtime')))
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS granted ON granted.oid=membership.roleid
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
WHERE member.rolname IN (
 'harpy_migrator','harpy_v2_runtime','harpy_adopter')`).Scan(&membershipsValid)
	if err != nil {
		t.Fatal(err)
	}
	err = database.admin.QueryRowContext(ctx, `SELECT
 pg_catalog.has_schema_privilege('harpy_runtime','platform','USAGE') AND
 NOT pg_catalog.has_schema_privilege('harpy_adopter','platform','USAGE')`).
		Scan(&schemaValid)
	if err != nil || !rolesValid || !membershipsValid || !schemaValid {
		t.Fatalf("role/membership/schema contract=%v/%v/%v err=%v",
			rolesValid, membershipsValid, schemaValid, err)
	}
}

func assertIAMPlatformTriggerCatalog(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var valid bool
	err := database.admin.QueryRowContext(ctx, `
SELECT count(*)=2 AND
 count(*) FILTER (WHERE trigger.tgname='iam_refresh_locators_guard' AND
  relation.relname='iam_refresh_locators' AND
  procedure.proname='iam_refresh_locator_guard_v1' AND
  trigger.tgenabled='A' AND trigger.tgtype=31)=1 AND
 count(*) FILTER (WHERE trigger.tgname='iam_auth_throttle_buckets_guard' AND
  relation.relname='iam_auth_throttle_buckets' AND
  procedure.proname='iam_auth_throttle_guard_v1' AND
  trigger.tgenabled='A' AND trigger.tgtype=31)=1
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation ON relation.oid=trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
JOIN pg_catalog.pg_proc AS procedure ON procedure.oid=trigger.tgfoid
WHERE NOT trigger.tgisinternal AND namespace.nspname='platform'
  AND relation.relname IN ('iam_refresh_locators','iam_auth_throttle_buckets')`).
		Scan(&valid)
	if err != nil || !valid {
		t.Fatalf("IAM platform ALWAYS trigger catalog invalid: %v", err)
	}
}

func assertIAMPlatformTableACL(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	for _, role := range []string{"harpy_runtime", "harpy_adopter"} {
		for _, table := range []string{
			"iam_login_directory", "iam_refresh_locators", "iam_auth_throttle_buckets",
		} {
			var privilege bool
			query := fmt.Sprintf(`SELECT
pg_catalog.has_table_privilege('%s','platform.%s','SELECT') OR
pg_catalog.has_table_privilege('%s','platform.%s','INSERT') OR
pg_catalog.has_table_privilege('%s','platform.%s','UPDATE') OR
pg_catalog.has_table_privilege('%s','platform.%s','DELETE')`,
				role, table, role, table, role, table, role, table)
			if err := database.admin.QueryRowContext(ctx, query).Scan(&privilege); err != nil {
				t.Fatal(err)
			}
			if privilege {
				t.Fatalf("%s has direct DML on platform.%s", role, table)
			}
		}
	}
}
