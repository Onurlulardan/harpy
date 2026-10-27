//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

const billingIsolationTenantID = "71000000-0000-4000-8000-000000000072"

type protectedBillingAuthority struct {
	scopeKind, scopeID, groupID, checkpoint string
	operationID, runID                      uuid.UUID
	planBytes                               []byte
	planSHA, leaseManifest, sourceXID       string
	finalLSN                                string
}

func newProtectedBillingAuthority(t *testing.T, scope string,
	plan billingv1adoptionapp.ScopePlan, operationID, runID, sourceXID string,
) protectedBillingAuthority {
	t.Helper()
	planBytes, planSHA, err := billingv1adoptionapp.EncodeScopePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	group, checkpoint := "billing_forms", "billing"
	if scope == "global" {
		group, checkpoint = "billing", "catalog"
	}
	return protectedBillingAuthority{
		scopeKind: scope, scopeID: plan.ScopeID, groupID: group,
		checkpoint: checkpoint, operationID: uuid.MustParse(operationID),
		runID: uuid.MustParse(runID), planBytes: planBytes, planSHA: planSHA,
		leaseManifest: fmt.Sprintf("%064x", len(sourceXID)+len(scope)+91),
		sourceXID:     sourceXID, finalLSN: "0/90",
	}
}

func prepareProtectedBillingAuthorities(t *testing.T, ctx context.Context,
	database *sql.DB, authorities ...protectedBillingAuthority,
) {
	t.Helper()
	prepareBillingCutoverSource(t, ctx, database)
	tenantID := uuid.MustParse(billingIsolationTenantID)
	if _, err := database.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,'t_billing_isolation','ACTIVE',NULL)`, tenantID); err != nil {
		t.Fatalf("seed Billing isolation tenant: %v", err)
	}
	fixture := newBillingCutoverAuthorityFixture(
		"tenant", tenantID, "billing_forms", 1,
		"shadow", "v1", nil, "billing", "forms",
	)
	fixture.reason = "billing isolation cutover"
	fixture.changeReference = "billing-cutover:isolation"
	if err := seedBillingCutoverAuthority(ctx, database, fixture); err != nil {
		t.Fatalf("seed Billing isolation authority: %v", err)
	}
	expiry := time.Now().UTC().Add(20 * time.Minute).Truncate(time.Microsecond)
	expiryLiteral := expiry.Format(time.RFC3339Nano)
	if _, err := database.ExecContext(ctx, fmt.Sprintf(`
REVOKE harpy_owner FROM harpy_adopter;
ALTER ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0
VALID UNTIL '%s' PASSWORD NULL;
ALTER ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0
VALID UNTIL '%s' PASSWORD NULL;`, expiryLiteral, expiryLiteral)); err != nil {
		t.Fatalf("seal Billing cutover roles: %v", err)
	}
	for _, authority := range authorities {
		seedProtectedBillingOperation(t, ctx, database, authority, expiry)
	}
}

func seedProtectedBillingOperation(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedBillingAuthority, expiry time.Time,
) {
	t.Helper()
	const operationSQL = `INSERT INTO platform.cutover_operations(
operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
source_system_identifier,source_database_oid,target_system_identifier,
target_database_oid,publication_name,slot_name,state,created_at,updated_at)
SELECT $1,$2,gen_random_uuid(),gen_random_uuid(),$3,$4,$5,$6,'billing',1,1,
gen_random_uuid(),$7,clock_timestamp(),$8,$9,$10,control.system_identifier,
db.oid,control.system_identifier,db.oid,$11,$12,'quiescing',
clock_timestamp(),clock_timestamp()
FROM pg_control_system() AS control
JOIN pg_database AS db ON db.datname=current_database()`
	transformSHA := fmt.Sprintf("%064x", len(a.planBytes)+37)
	publication := "billing_" + a.scopeKind + "_adopt_pub"
	slot := "billing_" + a.scopeKind + "_adopt_slot"
	execProtectedBillingSQL(t, ctx, database, operationSQL,
		a.operationID, a.runID, billingEnvironmentID, a.scopeKind, a.scopeID,
		a.groupID, expiry, a.leaseManifest, a.planSHA, transformSHA,
		publication, slot)
	seedProtectedBillingBindings(t, ctx, database, a)
}

func seedProtectedBillingBindings(t *testing.T, ctx context.Context,
	database *sql.DB, a protectedBillingAuthority,
) {
	t.Helper()
	procedureName := "billing_v1_adoption_" + a.scopeKind + "_apply_v1"
	identity := "uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn," +
		"xid8,integer,jsonb,text"
	signature := "platform." + procedureName + "(" + identity + ")"
	contractID := "billing.v1_adoption." + a.scopeKind + ".v1"
	execProtectedBillingSQL(t, ctx, database, `
INSERT INTO platform.cutover_operation_relations(
 operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
 source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
 target_schema_name,target_relation_name,replica_identity)
SELECT $1,$2,1,namespace.oid,source.oid,'platform','billing_v1_adoption_source',
 namespace.oid,target.oid,'platform','billing_v1_adoption_plan_chunks','primary_key'
FROM pg_namespace AS namespace
JOIN pg_class AS source ON source.relnamespace=namespace.oid
 AND source.relname='billing_v1_adoption_source'
JOIN pg_class AS target ON target.relnamespace=namespace.oid
 AND target.relname='billing_v1_adoption_plan_chunks'
WHERE namespace.nspname='platform'`, a.operationID, a.runID)
	execProtectedBillingSQL(t, ctx, database, `
INSERT INTO platform.cutover_operation_checkpoint_keys
(operation_id,run_id,relation_ordinal,ordinal,checkpoint_key)
VALUES ($1,$2,1,1,$3)`, a.operationID, a.runID, a.checkpoint)
	execProtectedBillingSQL(t, ctx, database, `
INSERT INTO platform.cutover_operation_procedures
(operation_id,run_id,ordinal,relation_ordinal,procedure_kind,procedure_oid,
 procedure_schema,procedure_name,identity_arguments,procedure_signature,
 procedure_contract_id)
VALUES ($1,$2,1,1,'apply',$3::regprocedure::oid,'platform',$4,$5,$3,$6)`,
		a.operationID, a.runID, signature, procedureName, identity, contractID)
	execProtectedBillingSQL(t, ctx, database, `
INSERT INTO platform.cutover_change_journal
(operation_id,run_id,scope_kind,scope_id,domain,cutover_generation,source_xid,
 source_session_user)
VALUES ($1,$2,$3,$4,'billing',1,$5::xid8,'harpy_v1_runtime')`,
		a.operationID, a.runID, a.scopeKind, a.scopeID, a.sourceXID)
}

func execProtectedBillingSQL(t *testing.T, ctx context.Context,
	database *sql.DB, statement string, arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed protected Billing authority: %v", err)
	}
}
