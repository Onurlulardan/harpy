//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const providerPGImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

const providerPGEnvironment = "70000000-0000-4000-8000-000000000007"
const providerPGTenant = "71000000-0000-4000-8000-000000000071"

func prepareProviderPG18(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, providerPGImage,
		tcpostgres.WithDatabase("harpy_provider_command"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("provider-command-owner-only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("start PostgreSQL 18.4 provider command database: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err = database.ExecContext(ctx, providerPGPrerequisites); err != nil {
		t.Fatalf("prepare provider command prerequisites: %v", err)
	}
	seedProviderPGCutoverAuthorities(t, ctx, database)
	prepareBillingContractRetirementFixture(t, ctx, database)
	applyProviderPGDocuments(t, ctx, database)
	return database
}

func seedProviderPGCutoverAuthorities(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	global := newBillingCutoverAuthorityFixture(
		"global", uuid.MustParse(providerPGEnvironment), "billing",
		1, "v2", "v2", billingAcceptanceA, "billing",
	)
	if err = seedBillingCutoverAuthority(ctx, transaction, global); err == nil {
		tenant := newBillingCutoverAuthorityFixture(
			"tenant", uuid.MustParse(providerPGTenant), "billing_forms",
			1, "v2", "v2", billingAcceptanceA, "billing", "forms",
		)
		err = seedBillingCutoverAuthority(ctx, transaction, tenant)
	}
	if err != nil {
		t.Fatalf("seed provider cutover authorities: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

const providerPGPrerequisites = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT PASSWORD 'runtime-only';
CREATE ROLE harpy_adopter LOGIN NOINHERIT PASSWORD 'adopter-only';
CREATE ROLE harpy_cutover_authority_owner NOLOGIN;
CREATE ROLE harpy_cutover_finalizer LOGIN NOINHERIT;
GRANT harpy_owner TO harpy_v2_runtime;
GRANT harpy_owner TO harpy_adopter;
GRANT harpy_cutover_authority_owner TO harpy_cutover_finalizer
    WITH INHERIT FALSE, SET TRUE;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_cutover_authority_owner;
CREATE TABLE platform.environment_identity(environment_id uuid PRIMARY KEY);
CREATE TABLE platform.tenants(
 id uuid PRIMARY KEY,schema_name name NOT NULL UNIQUE,status text NOT NULL,
 deleted_at timestamptz
);
CREATE TABLE platform.cutover_groups(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 generation bigint NOT NULL,state text NOT NULL,writer text NOT NULL,
 requested_by uuid NOT NULL,approved_by uuid,reason text NOT NULL,
 change_reference text NOT NULL,acceptance_digest text,
 changed_at timestamptz NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,group_id),
 UNIQUE(scope_kind,scope_id,group_id,generation)
);
CREATE TABLE platform.domain_cutovers(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,domain text NOT NULL,
 generation bigint NOT NULL,state text NOT NULL,writer text NOT NULL,
 requested_by uuid NOT NULL,approved_by uuid,reason text NOT NULL,
 change_reference text NOT NULL,acceptance_digest text,
 changed_at timestamptz NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,domain,generation)
);
CREATE TABLE platform.cutover_group_members(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 ordinal smallint NOT NULL,domain text NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,group_id,ordinal),
 FOREIGN KEY(scope_kind,scope_id,group_id)
  REFERENCES platform.cutover_groups(scope_kind,scope_id,group_id),
 FOREIGN KEY(scope_kind,scope_id,domain)
  REFERENCES platform.domain_cutovers(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,group_id,domain),
 UNIQUE(scope_kind,scope_id,domain)
);
CREATE TABLE platform.cutover_operations(
 operation_id uuid NOT NULL,run_id uuid NOT NULL,scope_kind text NOT NULL,
 scope_id uuid NOT NULL,group_id text NOT NULL,domain text NOT NULL,
 cutover_generation bigint NOT NULL,state text NOT NULL,
 PRIMARY KEY(operation_id,run_id)
);
INSERT INTO platform.environment_identity VALUES ('` + providerPGEnvironment + `');
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES ('` + providerPGTenant + `','t_provider','ACTIVE',NULL);`
