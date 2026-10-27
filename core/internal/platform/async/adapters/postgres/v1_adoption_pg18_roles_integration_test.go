//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const adoptionRoleGrants = `
CREATE ROLE harpy_adopter LOGIN PASSWORD 'adopter_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE harpy_cluster_admin LOGIN PASSWORD 'cluster_integration_only'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
GRANT USAGE ON SCHEMA platform TO harpy_adopter, harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION platform.platform_async_begin_v1_adoption_v1(
    uuid,text,uuid,text,text,bigint,bigint,bigint,bigint,text,text,
    smallint,smallint,smallint,bigint,timestamptz
) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_async_record_v1_inventory_item_v1(
    uuid,text,text,text,text,text,integer,text,text,text,text,integer,text,text,
    timestamptz,integer,integer,text,text,timestamptz,text
) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_async_seal_v1_inventory_v1(
    uuid,text
) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_async_adopt_v1_inventory_item_v1(
    uuid,text,text,text,text,uuid,timestamptz,text
) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_async_cancel_v1_inventory_item_v1(
    uuid,text,text,text,text,text,text,text,timestamptz,text
) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_async_record_v1_restart_fence_v1(
    uuid,text,text,text,text,text,text,timestamptz,timestamptz,text
) TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION platform.platform_async_record_v1_zero_inventory_v1(
    uuid,text,bigint,bigint,bigint,bigint,timestamptz,text
) TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION platform.platform_async_accept_v1_adoption_v1(
    uuid,text,text,timestamptz,text
) TO harpy_cluster_admin;`

func grantAdoptionRoles(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, adoptionRoleGrants); err != nil {
		t.Fatalf("grant adoption roles: %v", err)
	}
}
