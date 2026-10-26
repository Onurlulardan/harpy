//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

const (
	analyticsTenantA = "019b1000-0000-7000-8000-000000000001"
	analyticsTenantB = "019b1000-0000-7000-8000-000000000002"
)

func prepareAnalyticsPrerequisites(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	const roles = `
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD 'runtime-integration-only';
CREATE ROLE harpy_adopter LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD 'adopter-integration-only';
CREATE ROLE harpy_cluster_admin LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD 'admin-integration-only';
CREATE ROLE harpy_cutover_authority_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE;
CREATE ROLE harpy_cutover_finalizer LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE PASSWORD 'finalizer-integration-only';
GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT harpy_owner TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.tenants(id uuid PRIMARY KEY,schema_name name NOT NULL UNIQUE,status text NOT NULL,deleted_at timestamptz);
CREATE TABLE platform.artifact_quota_counters(
tenant_id uuid PRIMARY KEY REFERENCES platform.tenants(id) ON DELETE RESTRICT,
used_ciphertext_bytes bigint NOT NULL DEFAULT 0,
reserved_ciphertext_bytes bigint NOT NULL DEFAULT 0,
fenced boolean NOT NULL DEFAULT false,fence_code text,revision bigint NOT NULL DEFAULT 0,
updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
CHECK(used_ciphertext_bytes BETWEEN 0 AND 21474836480),
CHECK(reserved_ciphertext_bytes BETWEEN 0 AND 21474836480),
CHECK(used_ciphertext_bytes+reserved_ciphertext_bytes<=21474836480),
CHECK(fenced=(fence_code IS NOT NULL)));
CREATE TABLE platform.cutover_groups(scope_kind text,scope_id uuid,group_id text,generation bigint,state text,writer text,requested_by uuid,approved_by uuid,reason text,change_reference text,acceptance_digest text,changed_at timestamptz,PRIMARY KEY(scope_kind,scope_id,group_id));
CREATE TABLE platform.cutover_group_members(scope_kind text,scope_id uuid,group_id text,domain text,ordinal bigint NOT NULL,PRIMARY KEY(scope_kind,scope_id,group_id,domain));
CREATE TABLE platform.domain_cutovers(scope_kind text,scope_id uuid,domain text,generation bigint,state text,writer text,requested_by uuid,approved_by uuid,reason text,change_reference text,acceptance_digest text,changed_at timestamptz,PRIMARY KEY(scope_kind,scope_id,domain));
CREATE TABLE platform.catalog_package_owner_payloads(package_id uuid,version text,component_id uuid,owner text,contract text,content_sha256 text,document jsonb,release_id uuid,registered_at timestamptz);
CREATE SCHEMA t_analytics_a AUTHORIZATION harpy_owner;
CREATE SCHEMA t_analytics_b AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform,t_analytics_a,t_analytics_b TO harpy_runtime;
GRANT USAGE ON SCHEMA platform,t_analytics_a,t_analytics_b TO harpy_adopter;
GRANT USAGE ON SCHEMA platform TO harpy_cluster_admin,harpy_cutover_authority_owner;`
	if _, err := database.ExecContext(ctx, roles); err != nil {
		t.Fatalf("prepare Analytics roles and schemas: %v", err)
	}
	prepareAnalyticsAuthorityReconcileCapability(t, ctx, database)
	for _, tenant := range []struct{ id, schema string }{
		{analyticsTenantA, "t_analytics_a"}, {analyticsTenantB, "t_analytics_b"},
	} {
		prepareAnalyticsTenant(t, ctx, database, tenant.id, tenant.schema, origin)
	}
}

func prepareAnalyticsTenant(
	t *testing.T, ctx context.Context, database *sql.DB, tenantID, schema, origin string,
) {
	t.Helper()
	identityOrigin := "new"
	if origin == "v1_adopted" {
		identityOrigin = origin
	}
	statement := fmt.Sprintf(`
INSERT INTO platform.tenants VALUES ('%s','%s','ACTIVE',NULL);
INSERT INTO platform.cutover_groups(
scope_kind,scope_id,group_id,generation,state,writer,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at) VALUES(
'tenant','%s','analytics',1,'v2','v2',
'019b1000-0000-7000-8000-0000000000a1',
'019b1000-0000-7000-8000-0000000000a2','initial','analytics-initial',
'%s','2026-07-29T10:00:00Z');
INSERT INTO platform.cutover_group_members VALUES ('tenant','%s','analytics','analytics',1);
INSERT INTO platform.domain_cutovers(
scope_kind,scope_id,domain,generation,state,writer,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at) VALUES(
'tenant','%s','analytics',1,'v2','v2',
'019b1000-0000-7000-8000-0000000000a1',
'019b1000-0000-7000-8000-0000000000a2','initial','analytics-initial',
'%s','2026-07-29T10:00:00Z');
CREATE TABLE %s.tenancy_schema_identity(singleton boolean PRIMARY KEY CHECK(singleton),tenant_id uuid NOT NULL UNIQUE,schema_name text NOT NULL,origin text NOT NULL CHECK(origin IN('new','v1_adopted')),created_at timestamptz NOT NULL);
INSERT INTO %s.tenancy_schema_identity VALUES(true,'%s','%s','%s',clock_timestamp());`,
		tenantID, schema, tenantID, fmt.Sprintf("%064x", 1), tenantID, tenantID,
		fmt.Sprintf("%064x", 1), schema, schema, tenantID, schema, identityOrigin)
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare Analytics tenant %s: %v", schema, err)
	}
}
