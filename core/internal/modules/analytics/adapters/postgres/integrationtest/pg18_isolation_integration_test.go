//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsTenantIsolation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var authorities int
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM platform.analytics_tenant_authorities AS authority
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind=authority.scope_kind
 AND group_row.scope_id=authority.scope_id
 AND group_row.group_id=authority.authority_group_id
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind=authority.scope_kind
 AND domain_row.scope_id=authority.scope_id
 AND domain_row.domain=authority.authority_domain
WHERE authority.authority_mode='runtime_v2'
  AND authority.authority_group_generation=group_row.generation
  AND authority.authority_domain_generation=domain_row.generation
  AND group_row.state='v2' AND group_row.writer='v2'
  AND domain_row.state='v2' AND domain_row.writer='v2'`).Scan(&authorities); err != nil {
		t.Fatal(err)
	}
	if authorities != 2 {
		t.Fatalf("backfilled Analytics current authorities = %d, want 2", authorities)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			`RESET ROLE; RESET SESSION AUTHORIZATION`)
		_ = connection.Close()
	}()
	if _, err = connection.ExecContext(ctx, `SET SESSION AUTHORIZATION harpy_v2_runtime; SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	var queryID string
	err = transaction.QueryRowContext(ctx, `SELECT query_id::text FROM analytics_query_create_v1(
't_analytics_a','019b1000-0000-7000-8000-000000000021','019b1000-0000-7000-8000-000000000022',
'isolation-create',decode(repeat('01',32),'hex'),'019b1000-0000-7000-8000-000000000023',
'019b1000-0000-7000-8000-000000000024','Orders','{}'::bytea,'{}'::bytea,'{}'::bytea,clock_timestamp())`).Scan(&queryID)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("create tenant A Analytics query: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var tenantA, tenantB int
	if err = database.QueryRowContext(ctx, `SELECT count(*) FROM t_analytics_a.analytics_saved_queries`).Scan(&tenantA); err != nil {
		t.Fatal(err)
	}
	if err = database.QueryRowContext(ctx, `SELECT count(*) FROM t_analytics_b.analytics_saved_queries`).Scan(&tenantB); err != nil {
		t.Fatal(err)
	}
	if tenantA != 1 || tenantB != 0 || queryID == "" {
		t.Fatalf("Analytics tenant isolation counts = %d/%d, query=%q", tenantA, tenantB, queryID)
	}
}
