//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

const (
	analyticsSuccessorOperationA = "019c1000-0000-7000-8000-000000000901"
	analyticsSuccessorRunA       = "019c1000-0000-7000-8000-000000000902"
	analyticsSuccessorOperationB = "019c1000-0000-7000-8000-000000000903"
	analyticsSuccessorRunB       = "019c1000-0000-7000-8000-000000000904"
	analyticsSuccessorWrongOp    = "019c1000-0000-7000-8000-000000000905"
	analyticsSuccessorWrongRun   = "019c1000-0000-7000-8000-000000000906"
)

type analyticsSuccessorFixture struct {
	digest string
}

func prepareAnalyticsSuccessorPG18(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) analyticsSuccessorFixture {
	t.Helper()
	prepareAnalyticsPrerequisites(t, ctx, database, origin)
	applyAnalyticsCutoverDependencies(t, ctx, database)
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsDocuments(t, origin, "platform"))
	reconciled, remaining := reconcileAnalyticsAuthorities(t, ctx, database, true)
	if reconciled != 2 || remaining != 0 {
		t.Fatalf("initial Analytics authority reconcile = %d/%d", reconciled, remaining)
	}
	fixture := analyticsSuccessorFixture{digest: fmt.Sprintf("%064x", 901)}
	_, err := database.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=2,state='v2',writer='v2',
requested_by='019c1000-0000-7000-8000-000000000911',
approved_by='019c1000-0000-7000-8000-000000000912',
reason='finalized',change_reference='analytics-finalized',
acceptance_digest=$2,changed_at='2026-07-29T11:00:00Z'
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'`,
		analyticsTenantA, fixture.digest)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `UPDATE platform.domain_cutovers
SET generation=2,state='v2',writer='v2',
requested_by='019c1000-0000-7000-8000-000000000911',
approved_by='019c1000-0000-7000-8000-000000000912',
reason='finalized',change_reference='analytics-finalized',
acceptance_digest=$2,changed_at='2026-07-29T11:00:00Z'
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='analytics'`,
		analyticsTenantA, fixture.digest)
	if err != nil {
		t.Fatal(err)
	}
	seedAnalyticsSuccessorOperation(t, ctx, database,
		analyticsSuccessorOperationA, analyticsSuccessorRunA,
		"analytics", "analytics")
	createAnalyticsSuccessorCaller(t, ctx, database)
	return fixture
}

func seedAnalyticsSuccessorOperation(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, runID, groupID, domain string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.cutover_operations(
operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
source_system_identifier,source_database_oid,target_system_identifier,
target_database_oid,publication_name,slot_name,state,created_at,updated_at)
SELECT $1,$2,pg_catalog.uuidv7(),pg_catalog.uuidv7(),pg_catalog.uuidv7(),
'tenant',$3,$4,$5,1,1,pg_catalog.uuidv7(),clock_timestamp()+interval '1 hour',
clock_timestamp(),$6,$7,$8,control.system_identifier,database.oid,
control.system_identifier,database.oid,'analytics_successor_pub',
'analytics_successor_slot','completed',clock_timestamp(),clock_timestamp()
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()`,
		operationID, runID, analyticsTenantA, groupID, domain,
		fmt.Sprintf("%064x", 921), fmt.Sprintf("%064x", 922),
		fmt.Sprintf("%064x", 923))
	if err != nil {
		t.Fatalf("seed Analytics successor operation: %v", err)
	}
}

func seedAnalyticsWrongDomainSuccessor(
	t *testing.T, ctx context.Context, database *sql.DB, digest string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.cutover_groups(
scope_kind,scope_id,group_id,generation,state,writer,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at) VALUES(
'tenant',$1,'forms',2,'v2','v2',
'019c1000-0000-7000-8000-000000000911',
'019c1000-0000-7000-8000-000000000912','finalized','forms-finalized',$2,
'2026-07-29T11:00:00Z')`, analyticsTenantA, digest)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `INSERT INTO platform.domain_cutovers(
scope_kind,scope_id,domain,generation,state,writer,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at) VALUES(
'tenant',$1,'forms',2,'v2','v2',
'019c1000-0000-7000-8000-000000000911',
'019c1000-0000-7000-8000-000000000912','finalized','forms-finalized',$2,
'2026-07-29T11:00:00Z')`, analyticsTenantA, digest)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `INSERT INTO
platform.cutover_group_members VALUES('tenant',$1,'forms','forms',1)`,
		analyticsTenantA)
	if err != nil {
		t.Fatal(err)
	}
	seedAnalyticsSuccessorOperation(t, ctx, database,
		analyticsSuccessorWrongOp, analyticsSuccessorWrongRun, "forms", "forms")
}

func createAnalyticsSuccessorCaller(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `CREATE SCHEMA cutover_authority
AUTHORIZATION harpy_cutover_authority_owner;
GRANT USAGE ON SCHEMA cutover_authority TO harpy_cutover_finalizer;
CREATE FUNCTION cutover_authority.analytics_test_finalize_v1(uuid,uuid,text)
RETURNS uuid LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS 'SELECT platform.analytics_finalize_authority_successor_v1($1,$2,$3)';
ALTER FUNCTION cutover_authority.analytics_test_finalize_v1(uuid,uuid,text)
OWNER TO harpy_cutover_authority_owner;
REVOKE ALL ON FUNCTION
cutover_authority.analytics_test_finalize_v1(uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
cutover_authority.analytics_test_finalize_v1(uuid,uuid,text)
TO harpy_cutover_finalizer`)
	if err != nil {
		t.Fatal(err)
	}
}

func configureAnalyticsFinalizerConnection(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_cutover_finalizer`); err != nil {
		t.Fatal(err)
	}
}

func callAnalyticsSuccessor(
	ctx context.Context, connection *sql.Conn, operationID, runID, digest string,
) (string, error) {
	var authorityID string
	err := connection.QueryRowContext(ctx, `
SELECT cutover_authority.analytics_test_finalize_v1($1,$2,$3)`,
		operationID, runID, digest).Scan(&authorityID)
	return authorityID, err
}
