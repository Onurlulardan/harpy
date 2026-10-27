//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertACLAdoptionFunctions(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	allowed := map[string][]string{
		"harpy_adopter": {
			"platform.platform_async_begin_v1_adoption_v1(uuid,text,uuid,text,text,bigint,bigint,bigint,bigint,text,text,smallint,smallint,smallint,bigint,timestamptz)",
			"platform.platform_async_record_v1_inventory_item_v1(uuid,text,text,text,text,text,integer,text,text,text,text,integer,text,text,timestamptz,integer,integer,text,text,timestamptz,text)",
			"platform.platform_async_seal_v1_inventory_v1(uuid,text)",
			"platform.platform_async_adopt_v1_inventory_item_v1(uuid,text,text,text,text,uuid,timestamptz,text)",
			"platform.platform_async_cancel_v1_inventory_item_v1(uuid,text,text,text,text,text,text,text,timestamptz,text)",
		},
		"harpy_cluster_admin": {
			"platform.platform_async_record_v1_restart_fence_v1(uuid,text,text,text,text,text,text,timestamptz,timestamptz,text)",
			"platform.platform_async_record_v1_zero_inventory_v1(uuid,text,bigint,bigint,bigint,bigint,timestamptz,text)",
			"platform.platform_async_accept_v1_adoption_v1(uuid,text,text,timestamptz,text)",
		},
	}
	for role, functions := range allowed {
		for _, function := range functions {
			assertACLFunctionPrivilege(t, ctx, database, role, function, true)
			assertACLFunctionPrivilege(t, ctx, database, "harpy_runtime", function, false)
		}
	}
	assertACLFunctionPrivilege(t, ctx, database, "harpy_adopter",
		allowed["harpy_cluster_admin"][2], false)
	assertACLFunctionPrivilege(t, ctx, database, "harpy_cluster_admin",
		allowed["harpy_adopter"][0], false)
	for _, role := range []string{
		"harpy_runtime", "harpy_adopter", "harpy_cluster_admin",
	} {
		assertACLTablePrivilege(t, ctx, database, role,
			"platform.async_v1_adoption_runs", "SELECT", false)
		assertACLTablePrivilege(t, ctx, database, role,
			"platform.async_v1_adoption_runs", "UPDATE", false)
		assertACLTablePrivilege(t, ctx, database, role,
			"platform.async_v1_adoption_runs", "DELETE", false)
	}
}
