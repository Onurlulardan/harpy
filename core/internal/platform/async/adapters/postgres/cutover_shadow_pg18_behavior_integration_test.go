//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

const cutoverShadowOperation = "70000000-0000-4000-8000-000000000001"
const cutoverShadowRun = "70000000-0000-4000-8000-000000000002"
const cutoverShadowManifest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func assertCutoverShadowBehavior(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	adoptionID, sourceID, jobID := cutoverShadowUUIDs(t, ctx, database)
	seedCutoverShadowOperation(t, ctx, database)
	installCutoverShadowTestBridge(t, ctx, database)
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	setShadowSession(t, ctx, connection, "harpy_adopter")
	direct := `SELECT platform.platform_async_enqueue_documents_cutover_copy_v1(
        $1,$2,$3,$4,$5,pg_catalog.clock_timestamp())`
	_, err = connection.ExecContext(ctx, direct, cutoverShadowOperation,
		cutoverShadowRun, adoptionID, sourceID, jobID)
	requireShadowSQLState(t, err, "42501")
	bridge := `SELECT platform.test_documents_cutover_enqueue_v1(
        $1,$2,$3,$4,$5,pg_catalog.clock_timestamp())`
	for attempt := 0; attempt < 2; attempt++ {
		var persisted string
		if err := connection.QueryRowContext(ctx, bridge, cutoverShadowOperation,
			cutoverShadowRun, adoptionID, sourceID, jobID).Scan(&persisted); err != nil {
			t.Fatalf("exact enqueue attempt %d: %v", attempt+1, err)
		}
		if persisted != jobID {
			t.Fatalf("enqueued job = %s, want %s", persisted, jobID)
		}
	}
	_, tamperedSource, _ := cutoverShadowUUIDs(t, ctx, database)
	_, err = connection.ExecContext(ctx, bridge, cutoverShadowOperation,
		cutoverShadowRun, adoptionID, tamperedSource, jobID)
	requireShadowSQLState(t, err, "P5603")
	_, err = connection.ExecContext(ctx,
		"DELETE FROM platform.cutover_job_bindings WHERE job_id=$1", jobID)
	requireShadowSQLState(t, err, "42501")
	setShadowSession(t, ctx, connection, "harpy_v2_runtime")
	assertShadowWorkerAuthority(t, ctx, connection)
	assertShadowAuthorityListing(t, ctx, connection)
}

func cutoverShadowUUIDs(
	t *testing.T, ctx context.Context, database *sql.DB,
) (string, string, string) {
	t.Helper()
	var first, second, third string
	err := database.QueryRowContext(ctx,
		"SELECT uuidv7()::text,uuidv7()::text,uuidv7()::text").
		Scan(&first, &second, &third)
	if err != nil {
		t.Fatal(err)
	}
	return first, second, third
}

func seedCutoverShadowOperation(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `INSERT INTO platform.cutover_operations (
operation_id,run_id,scope_kind,scope_id,group_id,domain,cutover_generation,
lease_manifest_sha256,state,lease_expires_at) VALUES (
$1,$2,'tenant',$3,'documents','documents',1,$4,'copying',
pg_catalog.clock_timestamp()+interval '1 hour')`
	if _, err := database.ExecContext(ctx, statement, cutoverShadowOperation,
		cutoverShadowRun, asyncTenantA, cutoverShadowManifest); err != nil {
		t.Fatal(err)
	}
}

func installCutoverShadowTestBridge(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `CREATE FUNCTION platform.test_documents_cutover_enqueue_v1(
uuid,uuid,uuid,uuid,uuid,timestamptz) RETURNS uuid
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS 'SELECT platform.platform_async_enqueue_documents_cutover_copy_v1($1,$2,$3,$4,$5,$6)';
REVOKE ALL ON FUNCTION platform.test_documents_cutover_enqueue_v1(
uuid,uuid,uuid,uuid,uuid,timestamptz) FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.test_documents_cutover_enqueue_v1(
uuid,uuid,uuid,uuid,uuid,timestamptz) TO harpy_adopter;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

func setShadowSession(
	t *testing.T, ctx context.Context, connection *sql.Conn, role string,
) {
	t.Helper()
	if !strings.HasPrefix(role, "harpy_") {
		t.Fatalf("invalid integration role %q", role)
	}
	if _, err := connection.ExecContext(ctx,
		"SET SESSION AUTHORIZATION "+role); err != nil {
		t.Fatalf("set session authorization %s: %v", role, err)
	}
}

func assertShadowWorkerAuthority(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	statement := `SELECT platform.platform_async_assert_worker_authority_v1(
        'tenant',$1,'documents',1,'documents',1,'cutover_shadow')`
	if _, err := connection.ExecContext(ctx, statement, asyncTenantA); err != nil {
		t.Fatalf("current cutover shadow authority: %v", err)
	}
	_, err := connection.ExecContext(ctx, statement, asyncTenantB)
	requireShadowSQLState(t, err, "P5601")
}

func assertShadowAuthorityListing(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	rows, err := connection.QueryContext(ctx, listJobAuthoritiesSQL,
		false, "", asyncTenantA, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var scope, scopeID, group string
	var groupGeneration, generation int64
	var domain, mode string
	if !rows.Next() || rows.Scan(&scope, &scopeID, &group, &groupGeneration,
		&domain, &generation, &mode) != nil {
		t.Fatal("cutover shadow authority row is missing")
	}
	if scopeID != asyncTenantA || group != "documents" ||
		domain != "documents" || mode != "cutover_shadow" {
		t.Fatalf("cutover shadow authority = %s/%s/%s/%s", scopeID, group, domain, mode)
	}
}

func requireShadowSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("cutover shadow error = %v, want SQLSTATE %s", err, want)
	}
}
