//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func preparePipelineTenant(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
ALTER TABLE platform.tenants
    ADD COLUMN schema_name text,
    ADD COLUMN status text NOT NULL DEFAULT 'ACTIVE',
    ADD COLUMN deleted_at timestamptz;
UPDATE platform.tenants SET schema_name = CASE id
    WHEN '11111111-1111-4111-8111-111111111111'::uuid THEN 'pipeline_tenant_a'
    ELSE 'pipeline_tenant_b'
END;
ALTER TABLE platform.tenants ALTER COLUMN schema_name SET NOT NULL;
INSERT INTO platform.cutover_groups (
    scope_kind,scope_id,group_id,state,writer,generation
) VALUES
    ('tenant','11111111-1111-4111-8111-111111111111','workflow','v2','v2',1),
    ('tenant','22222222-2222-4222-8222-222222222222','workflow','v2','v2',1);
INSERT INTO platform.cutover_group_members (
    scope_kind,scope_id,group_id,domain
) VALUES
    ('tenant','11111111-1111-4111-8111-111111111111','workflow','workflow'),
    ('tenant','22222222-2222-4222-8222-222222222222','workflow','workflow');
CREATE SCHEMA pipeline_tenant_a AUTHORIZATION harpy_owner;
CREATE SCHEMA pipeline_tenant_b AUTHORIZATION harpy_owner;
CREATE TABLE pipeline_tenant_a.tenancy_schema_identity (
    singleton boolean PRIMARY KEY,
    tenant_id uuid NOT NULL,
    schema_name text NOT NULL,
    origin text NOT NULL,
    created_at timestamptz NOT NULL
);
CREATE TABLE pipeline_tenant_b.tenancy_schema_identity (
    singleton boolean PRIMARY KEY,
    tenant_id uuid NOT NULL,
    schema_name text NOT NULL,
    origin text NOT NULL,
    created_at timestamptz NOT NULL
);
INSERT INTO pipeline_tenant_a.tenancy_schema_identity VALUES (
    true,'11111111-1111-4111-8111-111111111111','pipeline_tenant_a',
    'new',pg_catalog.clock_timestamp()
);
INSERT INTO pipeline_tenant_b.tenancy_schema_identity VALUES (
    true,'22222222-2222-4222-8222-222222222222','pipeline_tenant_b',
    'new',pg_catalog.clock_timestamp()
);
CREATE TABLE pipeline_tenant_a.pipeline_effect_probe (
    event_id uuid PRIMARY KEY
);
GRANT USAGE ON SCHEMA pipeline_tenant_a, pipeline_tenant_b TO harpy_runtime;
GRANT SELECT ON pipeline_tenant_a.tenancy_schema_identity,
    pipeline_tenant_b.tenancy_schema_identity TO harpy_runtime;
GRANT SELECT, INSERT ON pipeline_tenant_a.pipeline_effect_probe TO harpy_runtime;
GRANT SELECT ON platform.tenants, platform.cutover_groups,
    platform.cutover_group_members, platform.domain_cutovers,
    platform.outbox, platform.inbox TO harpy_runtime;
GRANT UPDATE ON platform.tenants, platform.cutover_groups,
    platform.domain_cutovers TO harpy_runtime;
GRANT INSERT ON platform.inbox TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare pipeline tenant scope: %v", err)
	}
}
