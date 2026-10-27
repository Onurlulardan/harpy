//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func seedScopeRows(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	for _, tenant := range []struct {
		id, schema string
	}{{tenantA, "tenant_a"}, {tenantB, "tenant_b"}} {
		if _, err := admin.Exec(
			ctx, "INSERT INTO platform.tenants VALUES($1,$2,$3,'ACTIVE',NULL)",
			tenant.id, tenant.schema, tenant.schema,
		); err != nil {
			t.Fatal(err)
		}
		if _, err := admin.Exec(
			ctx, "INSERT INTO platform.cutover_groups VALUES('tenant',$1,'iam_system_resources','v2','v2',1)",
			tenant.id,
		); err != nil {
			t.Fatal(err)
		}
		if _, err := admin.Exec(
			ctx, "INSERT INTO platform.cutover_group_members VALUES('tenant',$1,'iam_system_resources',1,'iam'),('tenant',$1,'iam_system_resources',2,'system_resource_layout')",
			tenant.id,
		); err != nil {
			t.Fatal(err)
		}
		seedMultiGroupAuthorityRows(t, ctx, admin, tenant.id)
		if _, err := admin.Exec(
			ctx, "INSERT INTO platform.domain_cutovers VALUES('tenant',$1,'iam','v2','v2',1)",
			tenant.id,
		); err != nil {
			t.Fatal(err)
		}
		if _, err := admin.Exec(
			ctx, "INSERT INTO platform.domain_cutovers VALUES('tenant',$1,'system_resource_layout','v2','v2',1)",
			tenant.id,
		); err != nil {
			t.Fatal(err)
		}
		for _, domain := range []string{"crm"} {
			if _, err := admin.Exec(
				ctx, "INSERT INTO platform.domain_cutovers VALUES('tenant',$1,$2,'v2','v2',1)",
				tenant.id, domain,
			); err != nil {
				t.Fatal(err)
			}
		}
		statement := "INSERT INTO " + tenant.schema +
			".tenancy_schema_identity VALUES(true,$1,$2,'new',pg_catalog.now())"
		if _, err := admin.Exec(ctx, statement, tenant.id, tenant.schema); err != nil {
			t.Fatal(err)
		}
	}
	statements := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO platform.cutover_groups VALUES('global',$1,'billing','v2','v2',5)", []any{scopeEnvironmentID}},
		{"INSERT INTO platform.cutover_group_members VALUES('global',$1,'billing',1,'billing')", []any{scopeEnvironmentID}},
		{"INSERT INTO platform.domain_cutovers VALUES('global',$1,'billing','v2','v2',7)", []any{scopeEnvironmentID}},
		{"INSERT INTO platform.billing_global_probe VALUES('public-seed','visible')", nil},
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("seed global Billing scope: %v", err)
		}
	}
}
