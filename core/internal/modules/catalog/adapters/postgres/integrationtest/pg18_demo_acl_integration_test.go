//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func assertCatalogDemoOperationPlanACL(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	var selectAllowed, tableInsert, updateAllowed, deleteAllowed bool
	var insertColumns int
	err := database.owner.QueryRowContext(ctx, `SELECT
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_operation_plans','SELECT'),
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_operation_plans','INSERT'),
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_operation_plans','UPDATE'),
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_operation_plans','DELETE'),
(SELECT count(*) FROM information_schema.column_privileges
 WHERE grantee='harpy_runtime' AND table_schema='catalog_a'
 AND table_name='catalog_demo_operation_plans' AND privilege_type='INSERT')`).Scan(
		&selectAllowed, &tableInsert, &updateAllowed, &deleteAllowed, &insertColumns,
	)
	if err != nil || !selectAllowed || tableInsert || updateAllowed || deleteAllowed ||
		insertColumns != 8 {
		t.Fatalf("Catalog demo operation-plan ACL select/table-insert/update/delete/columns="+
			"%t/%t/%t/%t/%d err=%v", selectAllowed, tableInsert, updateAllowed,
			deleteAllowed, insertColumns, err)
	}
}
