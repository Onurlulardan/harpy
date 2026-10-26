//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertMaintenancePG18ACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const routines = `
SELECT pg_catalog.count(*),pg_catalog.bool_and(routine.prosecdef),
 pg_catalog.bool_and(owner.rolname='harpy_owner'),
 pg_catalog.bool_and(routine.proconfig=ARRAY[
  'search_path='||$1||', pg_catalog']::text[]),
	 pg_catalog.bool_and(CASE
	  WHEN routine.proname=ANY(ARRAY[
	   'iam_maintenance_profile_resources_read_v1',
	   'iam_maintenance_profile_resources_read_v2',
	   'iam_maintenance_profile_lineage_resources_read_v2',
	   'iam_maintenance_profile_schema_read_v1'])
	   THEN NOT pg_catalog.has_function_privilege(
	    'harpy_runtime',routine.oid,'EXECUTE')
	  ELSE pg_catalog.has_function_privilege(
	    'harpy_runtime',routine.oid,'EXECUTE') END),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'harpy_adopter',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'public',routine.oid,'EXECUTE'))
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
 'iam_maintenance_directory_repair_page_v1',
 'iam_maintenance_directory_checksum_sources_v1',
 'iam_maintenance_profile_page_read_v1',
	 'iam_maintenance_profile_mappings_read_v1',
	 'iam_maintenance_profile_schema_read_v1',
	 'iam_maintenance_profile_schema_read_v2',
	 'iam_maintenance_profile_resources_read_v1',
	 'iam_maintenance_profile_resources_read_v2',
	 'iam_maintenance_profile_lineage_resources_read_v2',
	 'iam_maintenance_profile_lineage_resources_read_v3'])`
	for _, tenant := range maintenancePG18Tenants {
		var count int
		var definer, owner, captured, runtime, adopter, public bool
		err := database.QueryRowContext(ctx, routines, tenant.schema).Scan(
			&count, &definer, &owner, &captured, &runtime, &adopter, &public)
		if err != nil || count != 10 || !definer || !owner || !captured ||
			!runtime || !adopter || !public {
			t.Fatalf("%s read ACL=%d/%t/%t/%t/%t/%t/%t: %v",
				tenant.schema, count, definer, owner, captured,
				runtime, adopter, public, err)
		}
		assertMaintenancePG18ProfileCommandACL(t, ctx, database, tenant.schema)
		assertMaintenancePG18ProfileCapabilityACL(t, ctx, database, tenant.schema)
		assertMaintenancePG18IdentityProfileACL(t, ctx, database, tenant.schema)
		assertMaintenancePG18NoTablePrivileges(t, ctx, database, tenant.schema)
	}
	assertMaintenancePG18DirectDMLDenied(t, ctx, database)
}

func assertMaintenancePG18ProfileCommandACL(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const statement = `
SELECT pg_catalog.count(*),pg_catalog.bool_and(routine.prosecdef),
 pg_catalog.bool_and(owner.rolname='harpy_owner'),
 pg_catalog.bool_and(routine.proconfig=ARRAY[
  'search_path='||$1||', pg_catalog']::text[]),
 pg_catalog.bool_and(CASE WHEN routine.proname=ANY(ARRAY[
  'iam_maintenance_profile_backfill_apply_v1',
	  'iam_maintenance_profile_backfill_apply_v2',
  'iam_maintenance_profile_backfill_finish_v1'])
  THEN NOT pg_catalog.has_function_privilege(
   'harpy_runtime',routine.oid,'EXECUTE')
  ELSE pg_catalog.has_function_privilege(
   'harpy_runtime',routine.oid,'EXECUTE') END),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'harpy_adopter',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'public',routine.oid,'EXECUTE'))
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
 'iam_maintenance_profile_backfill_apply_v1',
 'iam_maintenance_profile_backfill_apply_v2',
	 'iam_maintenance_profile_backfill_apply_v3',
 'iam_maintenance_profile_backfill_finish_v1',
 'iam_maintenance_profile_backfill_finish_v2'])`
	var count int
	var definer, owner, captured, runtime, adopter, public bool
	err := database.QueryRowContext(ctx, statement, schema).Scan(
		&count, &definer, &owner, &captured, &runtime, &adopter, &public)
	if err != nil || count != 5 || !definer || !owner || !captured ||
		!runtime || !adopter || !public {
		t.Fatalf("%s profile command ACL=%d/%t/%t/%t/%t/%t/%t: %v",
			schema, count, definer, owner, captured, runtime, adopter, public, err)
	}
}

func assertMaintenancePG18NoTablePrivileges(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	for _, table := range []string{
		"iam_user_identifiers", "iam_users", "iam_maintenance_checkpoints",
		"iam_profile_backfill_plans", "iam_profile_backfill_pages",
		"iam_profile_backfill_mappings", "iam_profile_schema_versions",
		"iam_profile_backfill_preconversions",
		"iam_profile_backfill_worker_capabilities",
	} {
		for _, role := range []string{"harpy_runtime", "harpy_adopter", "public"} {
			var forbidden bool
			err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_table_privilege($1,$2,'SELECT,INSERT,UPDATE,DELETE')`,
				role, schema+"."+table).Scan(&forbidden)
			if err != nil || forbidden {
				t.Fatalf("%s has direct %s.%s privilege: %v",
					role, schema, table, err)
			}
		}
	}
}

func assertMaintenancePG18DirectDMLDenied(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
		"SET LOCAL search_path=t_iam_a,platform,pg_catalog",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	_, err = transaction.ExecContext(ctx,
		"UPDATE iam_users SET revision=revision+1 WHERE id=$1", maintenanceUserOne)
	_ = transaction.Rollback()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "permission") {
		t.Fatalf("runtime direct maintenance DML error=%v", err)
	}
}
