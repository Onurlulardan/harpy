//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertCutoverFunctionAuthority(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	var count, exact int
	err := admin.QueryRow(ctx, `SELECT count(*), count(*) FILTER (
    WHERE owner.rolname='harpy_owner' AND NOT owner.rolsuper
      AND NOT owner.rolinherit AND NOT owner.rolcreatedb AND NOT owner.rolcreaterole
      AND NOT owner.rolreplication AND NOT owner.rolbypassrls
      AND routine.prosecdef AND routine.proconfig @> ARRAY['search_path=pg_catalog']::text[])
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_authid AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname='platform' AND routine.proname IN (
    'cutover_probe_apply_v1','cutover_prepare_batch_v1','cutover_record_batch_v1',
	'cutover_probe_apply_and_record_v1')`).Scan(&count, &exact)
	if err != nil || count != 4 || exact != 4 {
		t.Fatalf("cutover function owner contract = %d/%d, err=%v", exact, count, err)
	}
	var canSet bool
	err = admin.QueryRow(ctx, `SELECT pg_catalog.pg_has_role(
    'harpy_migrator','harpy_owner','SET')`).Scan(&canSet)
	if err != nil || !canSet {
		t.Fatalf("migrator owner SET membership = %t, err=%v", canSet, err)
	}
}
