//go:build integration

package main

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	workflowPG18Image = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
		"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"
	workflowDatabaseName = "harpy_authority"
	workflowPassword     = "authority-integration-secret"
)

type workflowDatabase struct {
	container           testcontainers.Container
	admin               *pgx.Conn
	dsn                 string
	target              cutoverv1.DatabaseIdentity
	finalizerValidUntil time.Time
}

func startWorkflowDatabase(t *testing.T, ctx context.Context) workflowDatabase {
	t.Helper()
	container, err := postgrescontainer.Run(ctx, workflowPG18Image,
		postgrescontainer.WithDatabase(workflowDatabaseName),
		postgrescontainer.WithUsername("postgres"),
		postgrescontainer.WithPassword(workflowPassword),
		postgrescontainer.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start exact PostgreSQL 18.4: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	finalizerValidUntil := prepareWorkflowRoles(t, ctx, admin)
	prepareWorkflowSchemas(t, ctx, admin)
	target := observeWorkflowIdentity(t, ctx, admin)
	return workflowDatabase{container: container, admin: admin, dsn: dsn,
		target: target, finalizerValidUntil: finalizerValidUntil}
}

func prepareWorkflowRoles(t *testing.T, ctx context.Context, admin *pgx.Conn) time.Time {
	t.Helper()
	expires := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	finalizerExpires := time.Now().UTC().Add(90 * time.Second).Truncate(time.Microsecond)
	statement := fmt.Sprintf(`CREATE ROLE harpy_owner NOLOGIN NOINHERIT;
CREATE ROLE harpy_cutover_authority_owner NOLOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
CONNECTION LIMIT -1 PASSWORD NULL VALID UNTIL 'infinity';
CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1
PASSWORD '%s' VALID UNTIL 'infinity';
GRANT harpy_owner TO harpy_migrator
WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_cluster_admin LOGIN NOINHERIT NOSUPERUSER CREATEDB
CREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD '%s' VALID UNTIL '%s';
CREATE ROLE harpy_extension_admin LOGIN NOINHERIT SUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD '%s' VALID UNTIL '%s';
CREATE ROLE harpy_cutover_finalizer LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD '%s' VALID UNTIL '%s';
ALTER DATABASE %s OWNER TO harpy_owner;
GRANT CONNECT ON DATABASE %s TO harpy_cluster_admin,harpy_cutover_finalizer;`, workflowPassword,
		workflowPassword, expires, workflowPassword, expires, workflowPassword,
		finalizerExpires.Format(time.RFC3339Nano), workflowDatabaseName, workflowDatabaseName)
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("prepare authority roles: %v", err)
	}
	return finalizerExpires
}

func observeWorkflowIdentity(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) cutoverv1.DatabaseIdentity {
	t.Helper()
	identity := cutoverv1.DatabaseIdentity{
		DatabaseName: workflowDatabaseName, DatabaseOrigin: "fresh",
	}
	err := admin.QueryRow(ctx, `SELECT control.system_identifier::text,database.oid
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()`).Scan(
		&identity.SystemIdentifier, &identity.DatabaseOID,
	)
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

func workflowRoleDSN(t *testing.T, dsn, role string) string {
	t.Helper()
	config, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.User = url.UserPassword(role, workflowPassword)
	if configured := config.User.Username(); configured != role {
		t.Fatalf("role DSN user differs: got %s want %s", configured, role)
	}
	return config.String()
}

func connectWorkflowRole(
	t *testing.T, ctx context.Context, database workflowDatabase, role string,
) *pgx.Conn {
	t.Helper()
	connection, err := pgx.Connect(ctx, workflowRoleDSN(t, database.dsn, role))
	if err != nil {
		t.Fatal(err)
	}
	var sessionUser, currentUser string
	if err := connection.QueryRow(ctx,
		"SELECT SESSION_USER,CURRENT_USER",
	).Scan(&sessionUser, &currentUser); err != nil ||
		sessionUser != role || currentUser != role {
		_ = connection.Close(context.Background())
		t.Fatalf("role connection identity differs: session=%s current=%s err=%v",
			sessionUser, currentUser, err)
	}
	return connection
}

func TestWorkflowRoleDSNReplacesOriginalUser(t *testing.T) {
	value := workflowRoleDSN(t,
		"postgres://postgres:old@127.0.0.1:5432/harpy?sslmode=disable",
		"harpy_cluster_admin",
	)
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.User.Username() != "harpy_cluster_admin" ||
		parsed.User.Username() == "postgres" {
		t.Fatalf("role DSN retained original user: %s", parsed.User.Username())
	}
}
