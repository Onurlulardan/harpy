//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareScopeTenancyRuntimeRoutines(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
) {
	t.Helper()
	statements := []string{
		"SET ROLE harpy_owner",
		`CREATE FUNCTION platform.tenancy_runtime_resolve_command_v1(p_tenant_id uuid)
RETURNS TABLE(schema_name text, slug text, status text, not_deleted boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
	IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
	   CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
	   p_tenant_id IS NULL THEN
		RAISE EXCEPTION USING ERRCODE = '42501',
			MESSAGE = 'Tenant runtime resolver denied';
	END IF;
	RETURN QUERY SELECT
		tenant.schema_name::text,
		tenant.slug::text,
		tenant.status::text,
		tenant.deleted_at IS NULL
	FROM platform.tenants AS tenant
	WHERE tenant.id = p_tenant_id
	FOR SHARE;
END;
$function$`,
		`CREATE FUNCTION platform.tenancy_runtime_resolve_identity_v3(p_tenant_id uuid)
RETURNS TABLE(schema_name name, slug text, status text, not_deleted boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
	IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
	   CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
	   p_tenant_id IS NULL THEN
		RAISE EXCEPTION USING ERRCODE = '42501',
			MESSAGE = 'Tenant runtime identity resolver denied';
	END IF;
	RETURN QUERY SELECT
		tenant.schema_name::name,
		tenant.slug::text,
		tenant.status::text,
		(tenant.deleted_at IS NULL)::boolean
	FROM platform.tenants AS tenant
	WHERE tenant.id = p_tenant_id;
END;
$function$`,
		"RESET ROLE",
		"REVOKE ALL ON FUNCTION platform.tenancy_runtime_resolve_command_v1(uuid) FROM PUBLIC",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_runtime_resolve_command_v1(uuid) TO harpy_runtime",
		"REVOKE ALL ON FUNCTION platform.tenancy_runtime_resolve_identity_v3(uuid) FROM PUBLIC",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_runtime_resolve_identity_v3(uuid) TO harpy_runtime",
	}
	executeScopeSetup(t, ctx, admin, statements)
}
