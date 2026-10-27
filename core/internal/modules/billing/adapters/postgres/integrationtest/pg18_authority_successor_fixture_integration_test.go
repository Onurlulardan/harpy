//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	billingAcceptanceC          = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	billingGlobalSuccessorOp    = "019c2000-0000-7000-8000-000000000901"
	billingGlobalSuccessorRun   = "019c2000-0000-7000-8000-000000000902"
	billingTenantSuccessorOp    = "019c2000-0000-7000-8000-000000000903"
	billingTenantSuccessorRun   = "019c2000-0000-7000-8000-000000000904"
	billingFormsSuccessorOp     = "019c2000-0000-7000-8000-000000000905"
	billingFormsSuccessorRun    = "019c2000-0000-7000-8000-000000000906"
	billingWrongSuccessorRun    = "019c2000-0000-7000-8000-000000000907"
	billingCompetingGlobalOp    = "019c2000-0000-7000-8000-000000000908"
	billingCompetingGlobalRun   = "019c2000-0000-7000-8000-000000000909"
	billingWrongSuccessorDigest = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
)

type billingSuccessorResult struct {
	authorityID string
	err         error
}

func prepareBillingSuccessorFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	setBillingAuthorityGeneration(
		t, ctx, database, billingEnvironmentID, "global", 3, billingAcceptanceC,
	)
	setBillingAuthorityGeneration(
		t, ctx, database, billingTenantID, "tenant", 3, billingAcceptanceC,
	)
	seedBillingSuccessorOperation(
		t, ctx, database, billingGlobalSuccessorOp, billingGlobalSuccessorRun,
		"global", billingEnvironmentID, "billing",
	)
	seedBillingSuccessorOperation(
		t, ctx, database, billingTenantSuccessorOp, billingTenantSuccessorRun,
		"tenant", billingTenantID, "billing",
	)
	createBillingSuccessorCaller(t, ctx, database)
}

func seedBillingSuccessorOperation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID string,
	runID string,
	scopeKind string,
	scopeID string,
	groupID string,
	domain string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
INSERT INTO platform.cutover_operations(
 operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
 group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
 last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
 source_system_identifier,source_database_oid,target_system_identifier,
 target_database_oid,publication_name,slot_name,state,blocked_reason,
 created_at,updated_at
)
SELECT $1,$2,pg_catalog.gen_random_uuid(),pg_catalog.gen_random_uuid(),$7,
 $3,$4,$5,$6,2,1,pg_catalog.gen_random_uuid(),
 pg_catalog.clock_timestamp()+interval '1 hour',pg_catalog.clock_timestamp(),
 repeat('a',64),repeat('b',64),repeat('c',64),
 control.system_identifier,database.oid,control.system_identifier,database.oid,
 'billing_successor_pub','billing_successor_slot','completed',NULL,
 pg_catalog.clock_timestamp(),pg_catalog.clock_timestamp()
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database
 ON database.datname=pg_catalog.current_database()`,
		operationID, runID, scopeKind, scopeID, groupID, domain,
		billingEnvironmentID,
	); err != nil {
		t.Fatalf("seed Billing successor operation: %v", err)
	}
}

func createBillingSuccessorCaller(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
CREATE SCHEMA cutover_authority
AUTHORIZATION harpy_cutover_authority_owner;
GRANT USAGE ON SCHEMA cutover_authority TO harpy_cutover_finalizer;
CREATE FUNCTION cutover_authority.billing_test_finalize_v1(
 uuid,uuid,text
) RETURNS uuid LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS 'SELECT platform.billing_finalize_authority_successor_v1($1,$2,$3)';
ALTER FUNCTION cutover_authority.billing_test_finalize_v1(
 uuid,uuid,text
) OWNER TO harpy_cutover_authority_owner;
REVOKE ALL ON FUNCTION cutover_authority.billing_test_finalize_v1(
 uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.billing_test_finalize_v1(
 uuid,uuid,text
) TO harpy_cutover_finalizer`); err != nil {
		t.Fatalf("create Billing successor caller: %v", err)
	}
}

func openBillingFinalizerConnection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_cutover_finalizer`); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	return connection
}

func closeBillingFinalizerConnection(connection *sql.Conn) {
	_, _ = connection.ExecContext(
		context.Background(), `RESET SESSION AUTHORIZATION`,
	)
	_ = connection.Close()
}

func callBillingSuccessor(
	ctx context.Context,
	connection *sql.Conn,
	operationID string,
	runID string,
	formsOperationID any,
	formsRunID any,
	digest string,
) (string, error) {
	var authorityID string
	err := connection.QueryRowContext(ctx, `
SELECT cutover_authority.billing_test_finalize_v1($1,$2,$3)`,
		operationID, runID, digest,
	).Scan(&authorityID)
	return authorityID, err
}
