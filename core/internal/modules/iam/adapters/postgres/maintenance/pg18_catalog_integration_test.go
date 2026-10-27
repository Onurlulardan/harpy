//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

func assertMaintenancePG18Catalog(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertMaintenancePG18PlatformCatalog(t, ctx, database)
	const statement = `
SELECT
 (SELECT pg_catalog.count(*) FROM pg_catalog.pg_class AS relation
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
  WHERE namespace.nspname=$1 AND relation.relname=ANY(ARRAY[
   'iam_maintenance_schedules','iam_maintenance_checkpoints',
   'iam_maintenance_receipts','iam_maintenance_audit','iam_maintenance_metrics',
   'iam_directory_key_evidence','iam_profile_backfill_plans',
   'iam_profile_backfill_mappings','iam_profile_backfill_pages',
	   'iam_profile_backfill_errors','iam_profile_backfill_preconversions',
	   'iam_profile_backfill_worker_capabilities'])),
 (SELECT pg_catalog.count(DISTINCT routine.proname)
  FROM pg_catalog.pg_proc AS routine
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
  WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
   'iam_maintenance_bootstrap_v1','iam_maintenance_claim_assert_v1',
   'iam_maintenance_receipt_read_v1','iam_maintenance_complete_recurring_v1',
   'iam_maintenance_bcrypt_retire_v1','iam_maintenance_identifier_retire_v1',
   'iam_maintenance_directory_tombstone_purge_v1',
   'iam_maintenance_security_artifact_purge_v1',
   'iam_maintenance_directory_repair_apply_v1',
   'iam_maintenance_directory_repair_finish_v1',
   'iam_maintenance_directory_checksum_finish_v1',
   'iam_maintenance_profile_backfill_apply_v1',
   'iam_maintenance_profile_backfill_apply_v2',
	   'iam_maintenance_profile_backfill_apply_v3',
   'iam_maintenance_profile_backfill_error_v1',
   'iam_maintenance_profile_backfill_receipt_v1',
   'iam_maintenance_profile_backfill_finish_v1',
	   'iam_maintenance_profile_backfill_finish_v2',
	   'iam_maintenance_profile_lineage_resources_read_v3',
	   'iam_profile_worker_capability_mint_v1',
	   'iam_profile_worker_capability_consume_v1',
	   'iam_profile_worker_capability_commit_v1',
	   'iam_guard_profile_backfill_source_write_v3',
   'iam_profile_backfill_activate_v2'])),
 (SELECT pg_catalog.count(*) FROM pg_catalog.pg_proc AS routine
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
  WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
   'iam_maintenance_bootstrap_v1','iam_maintenance_receipt_read_v1',
   'iam_maintenance_bcrypt_retire_v1','iam_maintenance_identifier_retire_v1',
   'iam_maintenance_directory_tombstone_purge_v1',
   'iam_maintenance_security_artifact_purge_v1',
   'iam_maintenance_directory_repair_apply_v1',
   'iam_maintenance_directory_repair_finish_v1',
   'iam_maintenance_directory_checksum_finish_v1',
   'iam_maintenance_profile_backfill_apply_v1',
   'iam_maintenance_profile_backfill_apply_v2',
	   'iam_maintenance_profile_backfill_apply_v3',
   'iam_maintenance_profile_backfill_error_v1',
   'iam_maintenance_profile_backfill_finish_v1',
   'iam_maintenance_profile_backfill_finish_v2'])
  AND pg_catalog.has_function_privilege('harpy_runtime',routine.oid,'EXECUTE')
  AND NOT pg_catalog.has_function_privilege(0::oid,routine.oid,'EXECUTE'))`
	for _, tenant := range maintenancePG18Tenants {
		var tables, routines, sealed int
		err := database.QueryRowContext(ctx, statement, tenant.schema).Scan(
			&tables, &routines, &sealed)
		if err != nil || tables != 12 || routines != 24 || sealed != 12 {
			t.Fatalf("%s catalog=%d/%d/%d: %v",
				tenant.schema, tables, routines, sealed, err)
		}
	}
}

func assertMaintenancePG18PlatformCatalog(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
SELECT pg_catalog.count(*),pg_catalog.bool_and(routine.prosecdef),
 pg_catalog.bool_and(owner.rolname='harpy_owner'),
 pg_catalog.bool_and(pg_catalog.has_function_privilege(
  'harpy_runtime',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  0::oid,routine.oid,'EXECUTE'))
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname='platform' AND routine.proname=ANY(ARRAY[
 'iam_maintenance_directory_apply_v1','iam_maintenance_directory_purge_v1',
 'iam_maintenance_directory_snapshot_v1'])`
	var count int
	var definer, owner, runtime, sealed bool
	err := database.QueryRowContext(ctx, statement).Scan(
		&count, &definer, &owner, &runtime, &sealed)
	if err != nil || count != 3 || !definer || !owner || !runtime || !sealed {
		t.Fatalf("platform catalog=%d/%t/%t/%t/%t: %v",
			count, definer, owner, runtime, sealed, err)
	}
}
