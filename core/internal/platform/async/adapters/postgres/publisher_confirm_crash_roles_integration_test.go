//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func preparePipelineRoles(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
ALTER ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
    NOCREATEROLE NOREPLICATION NOBYPASSRLS;
ALTER ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
    NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB
    NOCREATEROLE NOREPLICATION NOBYPASSRLS;
DO $body$
BEGIN
    EXECUTE pg_catalog.format(
        'REVOKE CONNECT, TEMPORARY ON DATABASE %I FROM PUBLIC',
        pg_catalog.current_database()
    );
    EXECUTE pg_catalog.format(
        'GRANT CONNECT ON DATABASE %I TO harpy_runtime, harpy_v2_runtime',
        pg_catalog.current_database()
    );
END;
$body$;
CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner;
SET ROLE harpy_extension_admin;
CREATE EXTENSION vector WITH SCHEMA platform_ext;
RESET ROLE;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT SELECT ON ALL TABLES IN SCHEMA platform TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_assert_current_authority_v1(
    text,uuid,text,bigint,text,bigint
) TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare pipeline runtime roles: %v", err)
	}
}

func assertPipelineRuntimeGrants(
	t *testing.T,
	ctx context.Context,
	runtime *pgxpool.Pool,
) {
	t.Helper()
	var current string
	var tenantRead, inboxWrite bool
	err := runtime.QueryRow(ctx, `
SELECT current_user,
       pg_catalog.has_table_privilege(current_user,'platform.tenants','SELECT'),
       pg_catalog.has_table_privilege(current_user,'platform.inbox','INSERT')`).Scan(
		&current, &tenantRead, &inboxWrite,
	)
	if err != nil || current != "harpy_runtime" || !tenantRead || !inboxWrite {
		t.Fatalf("pipeline runtime grants = %s tenant=%t inbox=%t: %v",
			current, tenantRead, inboxWrite, err)
	}
}
