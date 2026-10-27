//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
	"github.com/google/uuid"
)

type activationScopeFixture struct {
	operationID, environmentID, tenantID, ownerID uuid.UUID
	name, slug, schema, kind, purpose, manifest   string
}

func newActivationScopeFixture(t *testing.T, kind, purpose, slug string) activationScopeFixture {
	t.Helper()
	tenantID := uuid.New()
	schema, err := domain.SchemaNameFromUUID(tenantID.String())
	if err != nil {
		t.Fatal(err)
	}
	return activationScopeFixture{
		operationID: uuid.New(), environmentID: uuid.MustParse(testEnvironmentID),
		tenantID: tenantID, ownerID: uuid.New(), name: "Activation " + slug,
		slug: slug, schema: schema.String(), kind: kind, purpose: purpose,
		manifest: strings.Repeat("a", 64),
	}
}

func prepareActivationScopeDatabase(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	prepareActivationScopeDatabaseThrough(t, ctx, database, origin, 0)
}

func prepareActivationScopeDatabaseThrough(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
	maxVersion int64,
) {
	t.Helper()
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrationsThrough(t, ctx, database, origin, maxVersion)
	insertDeploymentPolicy(t, ctx, database, "saas")
	for _, statement := range activationScopeFixtureSQL {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("prepare activation scope fixture: %v", err)
		}
	}
	for _, migration := range []string{
		"20260722080300_create_cutover_initialize_tenant_v1",
		"20260728103001_create_cutover_initialize_tenant_v2",
		"20260729100003_create_cutover_initialize_tenant_v3",
	} {
		path := filepath.Join("..", "..", "..", "..", "platform", "cutover",
			"migrations", origin, migration+"."+origin+".sql")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = database.ExecContext(ctx, string(content)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(path), err)
		}
	}
}

func insertActivationOperation(
	t *testing.T, ctx context.Context, database *sql.DB, fixture activationScopeFixture,
) {
	t.Helper()
	instanceOwner := any(nil)
	scopeKind, scopeID := "tenant", fixture.tenantID
	if fixture.kind == "environment_initialize" {
		instanceOwner = uuid.New()
		scopeKind, scopeID = "environment", fixture.environmentID
	}
	now := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	_, err := database.ExecContext(ctx, `INSERT INTO platform.tenancy_onboarding_operations (
operation_id,environment_id,operation_kind,idempotency_key,request_sha256,
manifest_sha256,step_plan_sha256,deployment_mode,tenant_id,schema_name,
binding_purpose,first_user_id,root_organization_id,protected_admin_binding_id,
instance_owner_binding_id,offering_version_id,offering_revision,offering_sha256,
offering_effective_at,registry_binding_id,authority_scope_kind,authority_scope_id,
authority_generation,authority_state_sha256,authority_proposal_sha256,
state,revision,created_at,updated_at
) VALUES (
$1,$2,$3,$4,repeat('1',64),$5,repeat('2',64),'saas',$6,$7,$8,$9,$10,$11,$12,
$13,1,repeat('3',64),$14,$15,$16,$17,1,repeat('4',64),repeat('5',64),
'finalizing',4,$14,$14
)`, fixture.operationID, fixture.environmentID, fixture.kind,
		fixture.operationID.String(), fixture.manifest, fixture.tenantID,
		fixture.schema, fixture.purpose, fixture.ownerID, uuid.New(), uuid.New(),
		instanceOwner, uuid.New(), now, uuid.New(), scopeKind, scopeID)
	if err != nil {
		t.Fatalf("insert %s activation operation: %v", fixture.kind, err)
	}
}

func callActivationV4(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture activationScopeFixture, application string,
) (bool, error) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	if _, err := transaction.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,true)`, application); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT platform.tenancy_activation_initialize_v4(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, fixture.operationID,
		fixture.environmentID, fixture.tenantID, fixture.name, fixture.slug,
		fixture.ownerID, fixture.schema, fixture.purpose, fixture.manifest,
		fixture.kind).Scan(&replayed)
	if err != nil {
		_ = transaction.Rollback()
		return false, err
	}
	return replayed, transaction.Commit()
}

var activationScopeFixtureSQL = []string{
	`CREATE TABLE platform.domain_cutovers (
scope_kind text,scope_id uuid,domain text,state text,writer text,generation bigint,
requested_by uuid,approved_by uuid,reason text,change_reference text,
acceptance_digest text,changed_at timestamptz DEFAULT statement_timestamp(),
PRIMARY KEY(scope_kind,scope_id,domain))`,
	`CREATE TABLE platform.cutover_groups (
scope_kind text,scope_id uuid,group_id text,state text,writer text,generation bigint,
requested_by uuid,approved_by uuid,reason text,change_reference text,
acceptance_digest text,changed_at timestamptz DEFAULT statement_timestamp(),
PRIMARY KEY(scope_kind,scope_id,group_id))`,
	`CREATE TABLE platform.cutover_group_members (
scope_kind text,scope_id uuid,group_id text,ordinal smallint,domain text,
PRIMARY KEY(scope_kind,scope_id,group_id,ordinal))`,
	`CREATE TABLE platform.activation_authority_calls (
domain text,authority_id uuid,tenant_id uuid,created_at timestamptz,
PRIMARY KEY(domain,tenant_id))`,
	activationAuthorityStub("analytics"),
}

func activationAuthorityStub(authority string) string {
	return `CREATE FUNCTION platform.` + authority + `_register_tenant_authority_v1(
uuid,uuid,bigint,bigint,timestamptz) RETURNS uuid LANGUAGE plpgsql
SECURITY DEFINER SET search_path TO pg_catalog AS $body$
BEGIN
IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
OR SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='authority registration denied';
END IF;
INSERT INTO platform.activation_authority_calls(domain,authority_id,tenant_id,created_at)
VALUES ('` + authority + `',$1,$2,$5) ON CONFLICT DO NOTHING;
RETURN $1;
END;$body$`
}
