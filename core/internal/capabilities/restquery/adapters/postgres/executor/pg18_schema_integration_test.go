//go:build integration

package executor

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const pg18TenantA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const pg18TenantB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
const pg18Actor = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"

func pg18PrepareSchema(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE ROLE forgecrud_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE forgecrud_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE forgecrud_forms_record_reader NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE forgecrud_forms_record_writer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE forgecrud_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '" + pg18RuntimePassword + "'",
		"CREATE ROLE forgecrud_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"GRANT forgecrud_runtime TO forgecrud_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT forgecrud_forms_record_reader TO forgecrud_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT forgecrud_forms_record_writer TO forgecrud_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"ALTER DATABASE forgecrud_restquery OWNER TO forgecrud_owner",
		"REVOKE CONNECT, TEMPORARY ON DATABASE forgecrud_restquery FROM PUBLIC",
		"GRANT CONNECT ON DATABASE forgecrud_restquery TO forgecrud_runtime, forgecrud_v2_runtime",
		"CREATE SCHEMA platform AUTHORIZATION forgecrud_owner",
		"CREATE SCHEMA platform_ext AUTHORIZATION forgecrud_owner",
		"SET ROLE forgecrud_extension_admin",
		"CREATE EXTENSION vector WITH SCHEMA platform_ext",
		"RESET ROLE",
		"CREATE SCHEMA tenant_a AUTHORIZATION forgecrud_owner",
		"CREATE SCHEMA tenant_b AUTHORIZATION forgecrud_owner",
		"SET ROLE forgecrud_owner",
		`CREATE TABLE platform.tenants(
id uuid PRIMARY KEY,schema_name text NOT NULL,status text NOT NULL,deleted_at timestamptz)`,
		`CREATE TABLE platform.cutover_groups(
scope_kind text,scope_id uuid,group_id text,state text,writer text,generation bigint,
PRIMARY KEY(scope_kind,scope_id,group_id))`,
		`CREATE TABLE platform.cutover_group_members(
scope_kind text,scope_id uuid,group_id text,ordinal smallint,domain text,
PRIMARY KEY(scope_kind,scope_id,group_id,ordinal))`,
		`CREATE TABLE platform.domain_cutovers(
scope_kind text,scope_id uuid,domain text,state text,writer text,generation bigint,
PRIMARY KEY(scope_kind,scope_id,domain))`,
		pg18TenantIdentitySQL("tenant_a"),
		pg18TenantIdentitySQL("tenant_b"),
		pg18RecordTableSQL("tenant_a"),
		pg18RecordTableSQL("tenant_b"),
		"RESET ROLE",
	}
	pg18ExecuteSetup(t, ctx, admin, statements)
}

func pg18TenantIdentitySQL(schema string) string {
	return "CREATE TABLE " + schema + `.tenancy_schema_identity(
singleton boolean PRIMARY KEY,tenant_id uuid,schema_name text,origin text,
created_at timestamptz)`
}

func pg18RecordTableSQL(schema string) string {
	return "CREATE TABLE " + schema + `.frm_executor_fixture(
fld_id uuid PRIMARY KEY,fld_name text,fld_name_folded text,fld_metadata jsonb)`
}

func pg18ExecuteSetup(
	t *testing.T, ctx context.Context, admin *pgx.Conn, statements []string,
) {
	t.Helper()
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare PostgreSQL 18.4 REST Query database: %v", err)
		}
	}
}
