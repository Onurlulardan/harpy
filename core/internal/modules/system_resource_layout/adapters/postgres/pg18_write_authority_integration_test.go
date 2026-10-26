//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertLayoutWriteRoutineACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var createV3, updateV2, createV2, createV1, updateV1 bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_create_v3(name,uuid,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_create_v2(name,uuid,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_update_v2(name,uuid,bigint,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_create_v1(name,uuid,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_update_v1(name,uuid,bigint,text,bytea,timestamptz)',
    'EXECUTE'
)`).Scan(&createV3, &createV2, &updateV2, &createV1, &updateV1)
	if err != nil {
		t.Fatal(err)
	}
	if !createV3 || createV2 || !updateV2 || createV1 || updateV1 {
		t.Fatalf("layout write ACL create3/create2/update2/create1/update1 = %t/%t/%t/%t/%t",
			createV3, createV2, updateV2, createV1, updateV1)
	}
}
