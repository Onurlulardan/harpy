//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestIAMAuthorizationRuntimeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "authorization_runtime_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			prepareIAMAuthorizationAuthority(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMAuthorizationRuntime(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMAuthorizationRuntimeBinding(t, ctx, database, tenant.schema)
				assertIAMAuthorizationRuntime(t, ctx, database, tenant)
				assertIAMAuthorizationSnapshot(t, ctx, database, tenant)
			}
			assertIAMAuthorizationFence(t, ctx, database, iamFoundationTenants[0])
		})
	}
}

func prepareIAMAuthorizationAuthority(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';
ALTER TABLE platform.domain_cutovers
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2';`); err != nil {
		t.Fatalf("prepare IAM authorization authority columns: %v", err)
	}
	applyIAMPlatformDocuments(t, ctx, database, []iamFoundationDocument{{
		path: "20260716012020_create_iam_tenant_authority_assertion_v1." + origin + ".sql",
		content: readIAMPlatformAuthMigration(t, origin,
			"20260716012020_create_iam_tenant_authority_assertion_v1"),
	}})
	identityOrigin := "new"
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
CREATE TABLE tenancy_schema_identity (
    singleton boolean PRIMARY KEY CHECK (singleton),
    tenant_id uuid NOT NULL UNIQUE,
    schema_name text NOT NULL,
	origin text NOT NULL CHECK (origin IN ('new', 'v1_adopted')),
	created_at timestamptz NOT NULL
)`)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO tenancy_schema_identity (
    singleton, tenant_id, schema_name, origin, created_at
) VALUES (true, $1, $2, $3, '2026-07-16T00:00:00Z')`,
			tenant.id, tenant.schema, identityOrigin)
	}
}

func applyIAMAuthorizationRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0,
		len(iamAuthorizationRuntimeMigrationNames))
	for _, name := range iamAuthorizationRuntimeMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func seedIAMAuthorizationRuntimeBinding(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_users SET authorization_version=1, revision=2,
    updated_at='2026-07-16T00:10:00Z'
WHERE id='51000000-0000-4000-8000-000000000001';
INSERT INTO iam_user_permission_bindings (
    id, user_id, permission_id, permission_match_kind, state,
    revision, created_at, updated_at, activated_at, revoked_at
) VALUES (
    '94000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 'exact', 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z',
    '2026-07-16T00:10:00Z', NULL
);`)
}

func assertIAMAuthorizationRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	document, err := runIAMAtomicText(ctx, database, tenant.schema, `
SELECT pg_catalog.jsonb_build_object(
    'actor', evidence.actor_document,
    'bindings', evidence.bindings_document,
    'clock', evidence.authorization_clock
)::text
FROM iam_authorization_evidence_read_v1(
    pg_catalog.current_schema()::name,
    '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
) AS evidence`)
	if err != nil || !strings.Contains(document,
		"94000000-0000-4000-8000-000000000001") {
		t.Fatalf("runtime authorization evidence failed: %v / %s", err, document)
	}
	document, err = runIAMAtomicText(ctx, database, tenant.schema, `
SELECT pg_catalog.jsonb_build_object(
    'actor', evidence.actor_document,
    'bindings', evidence.bindings_document,
    'clock', evidence.authorization_clock
)::text
FROM iam_authorization_evidence_read_v2(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
) AS evidence`, tenant.id)
	if err != nil || !strings.Contains(document,
		"94000000-0000-4000-8000-000000000001") {
		t.Fatalf("fenced V2 authorization evidence failed: %v / %s", err, document)
	}
	foreign := "t_iam_a"
	if tenant.schema == foreign {
		foreign = "t_iam_b"
	}
	_, err = runIAMAtomicText(ctx, database, tenant.schema, `
SELECT actor_document::text FROM iam_authorization_evidence_read_v1(
    $1::name, '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
)`, foreign)
	if err == nil {
		t.Fatal("authorization evidence accepted a foreign tenant schema")
	}
	foreignTenant := iamFoundationTenants[0].id
	if tenant.id == foreignTenant {
		foreignTenant = iamFoundationTenants[1].id
	}
	_, err = runIAMAtomicText(ctx, database, tenant.schema, `
SELECT actor_document::text FROM iam_authorization_evidence_read_v2(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', 'iam.users', 'read'
)`, foreignTenant)
	if err == nil {
		t.Fatal("authorization evidence V2 accepted a foreign tenant identity")
	}
}
