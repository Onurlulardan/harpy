//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func prepareCutoverShadowCompileFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, statement := range []string{
		cutoverShadowRoles,
		cutoverShadowAuthorityTables,
		cutoverShadowAuthorityRows,
	} {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("prepare cutover shadow compile fixture: %v", err)
		}
	}
}

const cutoverShadowRoles = `
CREATE ROLE harpy_adopter NOLOGIN;
CREATE ROLE harpy_cluster_admin NOLOGIN;
CREATE ROLE harpy_migrator NOLOGIN;
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN PASSWORD 'integration-runtime-only';
GRANT harpy_runtime TO harpy_v2_runtime;`

const cutoverShadowAuthorityTables = `
CREATE SCHEMA platform;
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY
);
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    state text NOT NULL,
    writer text NOT NULL,
    generation bigint NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, domain)
);
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    state text NOT NULL,
    writer text NOT NULL,
    generation bigint NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id)
);
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id, domain)
);
CREATE TABLE platform.cutover_operations (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    cutover_generation bigint NOT NULL,
    lease_manifest_sha256 text NOT NULL,
    state text NOT NULL,
    lease_expires_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, run_id)
);
CREATE TABLE platform.schema_jobs (
    job_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    owner text NOT NULL,
    resource_kind text NOT NULL,
    plan_contract_version text NOT NULL,
    compiler_version text NOT NULL,
    state text NOT NULL,
    generation bigint NOT NULL,
    next_step_ordinal integer,
    next_executor text,
    available_at timestamptz NOT NULL,
    finished_at timestamptz,
    submission_authority_id uuid,
    submission_authority_group_id text,
    submission_group_generation bigint,
    submission_billing_generation bigint,
    submission_forms_generation bigint
);
CREATE TABLE platform.schema_job_steps (
    job_id uuid NOT NULL REFERENCES platform.schema_jobs (job_id),
    ordinal integer NOT NULL,
    executor_kind text NOT NULL,
    state text NOT NULL,
    PRIMARY KEY (job_id, ordinal)
);
CREATE FUNCTION platform.billing_resolve_current_tenant_authority_receipt_v1(uuid)
RETURNS TABLE (
    authority_id uuid, authority_group_id text,
    authority_group_generation bigint, billing_generation bigint,
    forms_generation bigint
)
LANGUAGE sql STABLE SET search_path=pg_catalog
AS 'SELECT NULL::uuid,NULL::text,NULL::bigint,NULL::bigint,NULL::bigint';
`

const cutoverShadowAuthorityRows = `
INSERT INTO platform.tenants (id) VALUES
    ('11111111-1111-4111-8111-111111111111'),
    ('22222222-2222-4222-8222-222222222222');
INSERT INTO platform.cutover_groups VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'documents','shadow','v1',1),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'documents','shadow','v1',1);
INSERT INTO platform.domain_cutovers VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'documents','shadow','v1',1),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'documents','shadow','v1',1);
INSERT INTO platform.cutover_group_members VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'documents','documents'),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'documents','documents');`
