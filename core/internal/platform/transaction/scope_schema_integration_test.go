//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

var tenantA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
var tenantB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"

func prepareScopeDatabase(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_reader NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_writer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '" + scopeRuntimePassword + "'",
		"CREATE ROLE harpy_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_reader TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_writer TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"ALTER DATABASE harpy_scope OWNER TO harpy_owner",
		"REVOKE CONNECT, TEMPORARY ON DATABASE harpy_scope FROM PUBLIC",
		"GRANT CONNECT ON DATABASE harpy_scope TO harpy_runtime, harpy_v2_runtime",
		"CREATE SCHEMA platform AUTHORIZATION harpy_owner",
		"CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner",
		"SET ROLE harpy_extension_admin",
		"CREATE EXTENSION vector WITH SCHEMA platform_ext",
		"RESET ROLE",
		"CREATE SCHEMA tenant_a AUTHORIZATION harpy_owner",
		"CREATE SCHEMA tenant_b AUTHORIZATION harpy_owner",
		"SET ROLE harpy_owner",
		"CREATE TABLE platform.tenants(id uuid PRIMARY KEY, schema_name text NOT NULL, slug text NOT NULL, status text NOT NULL, deleted_at timestamptz)",
		"CREATE TABLE platform.cutover_groups(scope_kind text, scope_id uuid, group_id text, state text, writer text, generation bigint, PRIMARY KEY(scope_kind,scope_id,group_id))",
		"CREATE TABLE platform.cutover_group_members(scope_kind text, scope_id uuid, group_id text, ordinal smallint, domain text, PRIMARY KEY(scope_kind,scope_id,group_id,ordinal))",
		"CREATE TABLE platform.domain_cutovers(scope_kind text, scope_id uuid, domain text, state text, writer text, generation bigint, PRIMARY KEY(scope_kind,scope_id,domain))",
		"CREATE TABLE platform.billing_global_probe(id text PRIMARY KEY, value text NOT NULL)",
		"CREATE TABLE tenant_a.tenancy_schema_identity(singleton boolean PRIMARY KEY, tenant_id uuid, schema_name text, origin text, created_at timestamptz)",
		"CREATE TABLE tenant_b.tenancy_schema_identity(singleton boolean PRIMARY KEY, tenant_id uuid, schema_name text, origin text, created_at timestamptz)",
		"CREATE TABLE tenant_a.iam_probe(id uuid PRIMARY KEY, value text NOT NULL)",
		"CREATE TABLE tenant_b.iam_probe(id uuid PRIMARY KEY, value text NOT NULL)",
		"CREATE TABLE tenant_a.forms_probe(value text NOT NULL)",
		"CREATE TABLE tenant_b.forms_probe(value text NOT NULL)",
		"CREATE TABLE tenant_a.crm_probe(value text NOT NULL)",
		"CREATE TABLE tenant_b.crm_probe(value text NOT NULL)",
		"INSERT INTO tenant_a.forms_probe VALUES('tenant-a-forms')",
		"INSERT INTO tenant_b.forms_probe VALUES('tenant-b-forms')",
		"INSERT INTO tenant_a.crm_probe VALUES('tenant-a-crm')",
		"INSERT INTO tenant_b.crm_probe VALUES('tenant-b-crm')",
		"RESET ROLE",
	}
	executeScopeSetup(t, ctx, admin, statements)
	prepareScopeTenancyRuntimeRoutines(t, ctx, admin)
	prepareScopeCutoverRuntimeRoutines(t, ctx, admin)
	prepareScopeFoundation(t, ctx, admin)
	seedScopeRows(t, ctx, admin)
	grants := []string{
		"GRANT USAGE ON SCHEMA platform, tenant_a, tenant_b TO harpy_runtime",
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA platform TO harpy_runtime",
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA tenant_a, tenant_b TO harpy_runtime",
	}
	executeScopeSetup(t, ctx, admin, grants)
}

func executeScopeSetup(t *testing.T, ctx context.Context, admin *pgx.Conn, statements []string) {
	t.Helper()
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare tenant scope database: %v", err)
		}
	}
}
