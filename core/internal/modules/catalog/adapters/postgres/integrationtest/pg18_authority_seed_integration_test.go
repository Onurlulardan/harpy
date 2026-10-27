//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

var catalogAuthorityGroups = []struct {
	id      string
	members []string
}{
	{"analytics", []string{"analytics"}},
	{"billing_forms", []string{"billing"}},
	{"catalog", []string{"catalog"}},
	{"documents", []string{"documents"}},
	{"iam_system_resources", []string{"iam", "system_resource_layout"}},
	{"workflow", []string{"workflow"}},
}

func seedCatalogTenantAuthority(
	t *testing.T, ctx context.Context, database catalogPG18, origin string,
) {
	t.Helper()
	tenantOrigin := "new"
	if origin == "v1_adopted" {
		tenantOrigin = origin
	}
	for _, tenant := range catalogTenants {
		_, err := database.owner.ExecContext(ctx, `INSERT INTO platform.tenants
(id,slug,schema_name,status,deleted_at) VALUES($1,$2,$3,'ACTIVE',NULL)`,
			tenant.id, tenant.slug, tenant.schema)
		if err != nil {
			t.Fatal(err)
		}
		_, err = database.owner.ExecContext(ctx, `INSERT INTO `+tenant.schema+
			`.tenancy_schema_identity VALUES(true,$1,$2,$3,clock_timestamp())`,
			tenant.id, tenant.schema, tenantOrigin)
		if err != nil {
			t.Fatal(err)
		}
		seedCatalogAuthorityRows(t, ctx, database, tenant.id, tenant.generation)
	}
	_, err := database.owner.ExecContext(ctx, `GRANT SELECT ON
platform.tenants, platform.cutover_groups, platform.cutover_group_members,
platform.domain_cutovers TO harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
}

func seedCatalogAuthorityRows(t *testing.T, ctx context.Context,
	database catalogPG18, tenantID string, generation int64,
) {
	t.Helper()
	for _, group := range catalogAuthorityGroups {
		_, err := database.owner.ExecContext(ctx, `INSERT INTO platform.cutover_groups
(scope_kind,scope_id,group_id,generation,state,writer)
VALUES('tenant',$1,$2,$3,'v2','v2')`, tenantID, group.id, generation)
		if err != nil {
			t.Fatal(err)
		}
		for index, domain := range group.members {
			_, err = database.owner.ExecContext(ctx, `INSERT INTO platform.cutover_group_members
(scope_kind,scope_id,group_id,ordinal,domain) VALUES('tenant',$1,$2,$3,$4)`,
				tenantID, group.id, index+1, domain)
			if err != nil {
				t.Fatal(err)
			}
			_, err = database.owner.ExecContext(ctx, `INSERT INTO platform.domain_cutovers
(scope_kind,scope_id,domain,generation,state,writer)
VALUES('tenant',$1,$2,$3,'v2','v2')`, tenantID, domain, generation)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
