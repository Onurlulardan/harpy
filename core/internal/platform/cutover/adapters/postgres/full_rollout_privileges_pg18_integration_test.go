//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertFullRolloutPrivileges(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	var clusterExecute, runtimeExecute, clusterInsert, runtimeSelect, oldMissing bool
	err := admin.QueryRow(ctx, `SELECT
pg_catalog.has_function_privilege(
    'harpy_cluster_admin',
    'cutover_authority.record_full_rollout_completion_v1(uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,text,timestamptz,timestamptz,timestamptz)'::regprocedure,
    'EXECUTE'),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'cutover_authority.record_full_rollout_completion_v1(uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,text,timestamptz,timestamptz,timestamptz)'::regprocedure,
    'EXECUTE'),
pg_catalog.has_table_privilege(
    'harpy_cluster_admin',
    'cutover_authority.full_rollout_completion_receipts', 'INSERT'),
pg_catalog.has_table_privilege(
    'harpy_runtime',
    'cutover_authority.full_rollout_completion_receipts', 'SELECT'),
pg_catalog.to_regprocedure(
    'cutover_authority.record_full_rollout_completion_v1(uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,timestamptz,timestamptz,timestamptz)'
) IS NULL`).Scan(
		&clusterExecute, &runtimeExecute, &clusterInsert, &runtimeSelect, &oldMissing,
	)
	if err != nil {
		t.Fatalf("read full rollout privileges: %v", err)
	}
	if !clusterExecute || runtimeExecute || clusterInsert || runtimeSelect || !oldMissing {
		t.Fatalf("full rollout privileges = %v/%v/%v/%v old-missing=%v",
			clusterExecute, runtimeExecute, clusterInsert, runtimeSelect, oldMissing)
	}
}
