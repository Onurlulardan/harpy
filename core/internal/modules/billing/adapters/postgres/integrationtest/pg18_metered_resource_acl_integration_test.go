//go:build integration

package integrationtest

import "testing"

func TestMeteredResourceRuntimeHasOnlyTypedRoutineAccessOnPG18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)
			var tableCount int
			var hasDirectAccess bool
			err := database.QueryRowContext(ctx, `
SELECT count(*),bool_or(
    pg_catalog.has_table_privilege(
        'harpy_runtime',table_name,'SELECT'
    ) OR pg_catalog.has_table_privilege(
        'harpy_runtime',table_name,'INSERT'
    ) OR pg_catalog.has_table_privilege(
        'harpy_runtime',table_name,'UPDATE'
    ) OR pg_catalog.has_table_privilege(
        'harpy_runtime',table_name,'DELETE'
    )
)
FROM (VALUES
    ('platform.billing_metered_resource_receipts'),
    ('platform.billing_metered_resources'),
    ('platform.billing_metered_resource_reconcile_receipts')
) AS resource_tables(table_name)
`).Scan(&tableCount, &hasDirectAccess)
			if err != nil {
				t.Fatal(err)
			}
			if tableCount != 3 || hasDirectAccess {
				t.Fatalf("tables=%d direct_access=%v", tableCount, hasDirectAccess)
			}

			var routineCount, executableCount int
			err = database.QueryRowContext(ctx, `
SELECT count(*),count(*) FILTER (WHERE pg_catalog.has_function_privilege(
    'harpy_runtime',procedure_row.oid,'EXECUTE'
))
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace_row
  ON namespace_row.oid=procedure_row.pronamespace
WHERE namespace_row.nspname='platform'
  AND procedure_row.proname IN (
      'billing_metered_resource_inspect_v1',
      'billing_metered_resource_feature_v1',
      'billing_metered_resource_record_v1',
      'billing_metered_resource_reconcile_inspect_v1',
      'billing_metered_resource_reconcile_record_v1'
  )
`).Scan(&routineCount, &executableCount)
			if err != nil {
				t.Fatal(err)
			}
			if routineCount != 5 || executableCount != routineCount {
				t.Fatalf("routines=%d executable=%d", routineCount, executableCount)
			}
		})
	}
}
