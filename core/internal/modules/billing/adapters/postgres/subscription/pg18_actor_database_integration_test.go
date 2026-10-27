//go:build integration

package subscription

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const subscriptionPGImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

type subscriptionPGDatabase struct {
	owner   *sql.DB
	runtime *sql.DB
}

func startSubscriptionPG18(
	t *testing.T,
	ctx context.Context,
) subscriptionPGDatabase {
	t.Helper()
	container, err := tcpostgres.Run(
		ctx, subscriptionPGImage,
		tcpostgres.WithDatabase("harpy_billing_subscription"),
		tcpostgres.WithUsername("harpy_owner"),
		tcpostgres.WithPassword("billing-subscription-owner-only"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("start subscription PostgreSQL 18.4: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	ownerDSN, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := sql.Open("pgx", ownerDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	prepareSubscriptionPlatform(t, ctx, owner)
	applySubscriptionMigrations(t, ctx, owner)
	runtimeURL, err := url.Parse(ownerDSN)
	if err != nil {
		t.Fatal(err)
	}
	runtimeURL.User = url.UserPassword(
		"harpy_v2_runtime", "runtime-integration-only",
	)
	runtime, err := sql.Open("pgx", runtimeURL.String())
	if err != nil {
		t.Fatal(err)
	}
	runtime.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = runtime.Close() })
	if err := runtime.PingContext(ctx); err != nil {
		t.Fatalf("connect subscription runtime: %v", err)
	}
	return subscriptionPGDatabase{owner: owner, runtime: runtime}
}

func prepareSubscriptionPlatform(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT PASSWORD 'runtime-integration-only';
CREATE ROLE harpy_adopter LOGIN NOINHERIT PASSWORD 'adopter-integration-only';
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
 PRIMARY KEY(scope_kind,scope_id,group_id)
);
CREATE TABLE platform.cutover_group_members(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 ordinal smallint NOT NULL,domain text NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,group_id,domain)
);
CREATE TABLE platform.domain_cutovers(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,domain text NOT NULL,
 generation bigint NOT NULL,state text NOT NULL,writer text NOT NULL,
 PRIMARY KEY(scope_kind,scope_id,domain)
);
CREATE TABLE platform.cutover_operations(
 operation_id uuid NOT NULL,run_id uuid NOT NULL,scope_kind text NOT NULL,
 scope_id uuid NOT NULL,group_id text NOT NULL,domain text NOT NULL,
 cutover_generation bigint NOT NULL,state text NOT NULL,
 PRIMARY KEY(operation_id,run_id)
);
CREATE TABLE platform.jobs(
 id uuid PRIMARY KEY,tenant_id uuid,scope_kind text,scope_id uuid,
 authority_group_id text,authority_group_generation bigint,
 authority_domain text,cutover_generation bigint,payload_revision integer,
 payload_ref text,payload_sha256 text,state text
);
CREATE TABLE platform.test_subscription_events(event_id uuid PRIMARY KEY);
INSERT INTO platform.environment_identity VALUES('70000000-0000-4000-8000-000000000007');
INSERT INTO platform.tenants VALUES(
 '71000000-0000-4000-8000-000000000071','t_billing','ACTIVE',NULL
);
INSERT INTO platform.cutover_groups VALUES
 ('global','70000000-0000-4000-8000-000000000007','billing',1,'v2','v2'),
 ('tenant','71000000-0000-4000-8000-000000000071','billing_forms',1,'v2','v2');
INSERT INTO platform.cutover_group_members VALUES
 ('global','70000000-0000-4000-8000-000000000007','billing',1,'billing'),
 ('tenant','71000000-0000-4000-8000-000000000071','billing_forms',1,'billing'),
 ('tenant','71000000-0000-4000-8000-000000000071','billing_forms',2,'forms');
INSERT INTO platform.domain_cutovers VALUES
 ('global','70000000-0000-4000-8000-000000000007','billing',1,'v2','v2'),
 ('tenant','71000000-0000-4000-8000-000000000071','billing',1,'v2','v2'),
 ('tenant','71000000-0000-4000-8000-000000000071','forms',1,'v2','v2');`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare subscription platform: %v", err)
	}
	prepareSubscriptionCutoverFixture(t, ctx, database)
}

func applySubscriptionMigrations(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	working, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(working, "..", "..", "..", "migrations", "platform", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) != 202 {
		t.Fatalf("subscription migration count = %d, want 202", len(names))
	}
	for _, name := range names {
		document, readErr := os.ReadFile(filepath.Join(directory, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, applyErr := database.ExecContext(ctx, string(document)); applyErr != nil {
			t.Fatalf("apply %s: %v", name, applyErr)
		}
	}
}
