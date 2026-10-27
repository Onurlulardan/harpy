//go:build integration

package executor

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func pg18SeedTenants(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	for _, tenant := range []struct {
		id, schema string
	}{{pg18TenantA, "tenant_a"}, {pg18TenantB, "tenant_b"}} {
		pg18SeedTenantAuthority(t, ctx, admin, tenant.id, tenant.schema)
	}
	pg18SeedRecordRows(t, ctx, admin)
	pg18ExecuteSetup(t, ctx, admin, []string{
		"GRANT USAGE ON SCHEMA platform,tenant_a,tenant_b TO harpy_runtime",
		"GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA platform TO harpy_runtime",
		"GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA tenant_a,tenant_b TO harpy_runtime",
	})
}

func pg18SeedTenantAuthority(
	t *testing.T, ctx context.Context, admin *pgx.Conn, tenant, schema string,
) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO platform.tenants VALUES($1,$2,'ACTIVE',NULL)", []any{tenant, schema}},
		{`INSERT INTO platform.domain_cutovers
VALUES('tenant',$1,'forms','v2','v2',1)`, []any{tenant}},
		{`INSERT INTO platform.domain_cutovers
VALUES('tenant',$1,'crm','v2','v2',1)`, []any{tenant}},
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	identitySQL := "INSERT INTO " + schema +
		".tenancy_schema_identity VALUES(true,$1,$2,'new',pg_catalog.now())"
	if _, err := admin.Exec(ctx, identitySQL, tenant, schema); err != nil {
		t.Fatal(err)
	}
}

func pg18SeedRecordRows(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	for _, statement := range []string{
		`INSERT INTO tenant_a.frm_executor_fixture VALUES
('40000000-0000-4000-8000-000000000001','Ada','ada',NULL),
('40000000-0000-4000-8000-000000000002','Bora','bora','{"level":2,"active":true}'),
('40000000-0000-4000-8000-000000000003','Cem','cem','{"level":3}')`,
		`INSERT INTO tenant_b.frm_executor_fixture VALUES
('50000000-0000-4000-8000-000000000001','Tenant B One','tenant b one',NULL),
('50000000-0000-4000-8000-000000000002','Tenant B Two','tenant b two','{"tenant":"b"}')`,
	} {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
}
