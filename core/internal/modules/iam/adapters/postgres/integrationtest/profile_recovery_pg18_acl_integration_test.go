//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func assertProfileRecoveryACL(
	t *testing.T,
	ctx context.Context,
	db publicationDatabase,
) {
	t.Helper()
	var oldRead, oldWrite, oldInventoryAbsent, oldInventoryExecute bool
	var inventoryOpen, readV2, writeV2, directView bool
	err := db.admin.QueryRow(ctx, `SELECT
pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_read_contract_v1(name,text)',$1::text),'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_write_contract_read_v1(name,text)',$1::text),'EXECUTE'),
pg_catalog.to_regprocedure(pg_catalog.format(
 '%I.iam_profile_recovery_inventory_read_v1(name,uuid,uuid,integer,uuid,uuid,bigint,bigint)',
 $1::text)) IS NULL,
COALESCE(pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.to_regprocedure(pg_catalog.format(
  '%I.iam_profile_recovery_inventory_read_v1(name,uuid,uuid,integer,uuid,uuid,bigint,bigint)',
  $1::text)),'EXECUTE'),false),
pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_recovery_inventory_open_v1(name,uuid,uuid,uuid,bigint,bigint)',
  $1::text),'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_read_contract_v2(name,text)',$1::text),'EXECUTE'),
pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_write_contract_read_v2(name,text)',$1::text),'EXECUTE'),
pg_catalog.has_table_privilege(
 'harpy_runtime',pg_catalog.format(
  '%I.iam_profile_recovery_resources_v1',$1::text),'SELECT')`, db.tenant.schema).
		Scan(&oldRead, &oldWrite, &oldInventoryAbsent, &oldInventoryExecute,
			&inventoryOpen, &readV2, &writeV2, &directView)
	if err != nil {
		t.Fatalf("read profile recovery ACL: %v", err)
	}
	if oldRead || oldWrite || !oldInventoryAbsent || oldInventoryExecute ||
		!inventoryOpen || !readV2 || !writeV2 || directView {
		t.Fatalf(
			"profile recovery ACL old=%t/%t inventory=%t/%t/%t new=%t/%t view=%t",
			oldRead, oldWrite, oldInventoryAbsent, oldInventoryExecute,
			inventoryOpen, readV2, writeV2, directView,
		)
	}
}
