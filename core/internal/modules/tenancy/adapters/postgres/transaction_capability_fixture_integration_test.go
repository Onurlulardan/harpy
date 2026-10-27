//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type transactionCapabilityPlan struct {
	activationScopeFixture
	attemptID, migrationReceiptID, registryBindingID uuid.UUID
	leaseToken, stepPlan                             string
	revision                                         int64
}

func prepareTransactionCapabilityDatabase(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	prepareTransactionCapabilityDatabaseThrough(t, ctx, database, origin, 0)
}

func prepareTransactionCapabilityDatabaseThrough(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
	maxVersion int64,
) {
	t.Helper()
	prepareActivationScopeDatabaseThrough(t, ctx, database, origin, maxVersion)
	installCapabilityTestConsumers(t, ctx, database)
}

func newTransactionCapabilityPlan(t *testing.T, origin string) transactionCapabilityPlan {
	t.Helper()
	return transactionCapabilityPlan{
		activationScopeFixture: newActivationScopeFixture(
			t, "environment_initialize", "saas_control", "capability-"+origin,
		),
		attemptID: uuid.New(), migrationReceiptID: uuid.New(),
		registryBindingID: uuid.New(), leaseToken: strings.Repeat("9", 64),
		stepPlan: strings.Repeat("2", 64), revision: 4,
	}
}

func insertTransactionCapabilityPlan(
	t *testing.T, ctx context.Context, database *sql.DB, plan transactionCapabilityPlan,
) {
	t.Helper()
	now := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	execute := func(statement string, arguments ...any) {
		t.Helper()
		if _, execErr := transaction.ExecContext(
			ctx, statement, arguments...,
		); execErr != nil {
			t.Fatalf("insert transaction capability plan: %v", execErr)
		}
	}
	execute(`INSERT INTO platform.tenancy_onboarding_operations (
operation_id,environment_id,operation_kind,idempotency_key,request_sha256,
manifest_sha256,step_plan_sha256,deployment_mode,tenant_id,schema_name,
binding_purpose,first_user_id,root_organization_id,protected_admin_binding_id,
instance_owner_binding_id,offering_version_id,offering_revision,offering_sha256,
offering_effective_at,registry_binding_id,authority_scope_kind,authority_scope_id,
authority_generation,authority_state_sha256,authority_proposal_sha256,
state,revision,created_at,updated_at
) VALUES (
$1,$2,'environment_initialize',$3,repeat('1',64),$4,$5,'saas',$6,$7,
'saas_control',$8,$9,$10,$11,$12,1,repeat('3',64),$13,$14,'environment',$2,
1,repeat('4',64),repeat('5',64),'authority_acknowledged',$15,$13,$13
 )`,
		plan.operationID, plan.environmentID, plan.operationID.String(),
		plan.manifest, plan.stepPlan, plan.tenantID, plan.schema, plan.ownerID,
		uuid.New(), uuid.New(), uuid.New(), uuid.New(), now,
		plan.registryBindingID, plan.revision)
	execute(`INSERT INTO platform.tenancy_onboarding_steps(
operation_id,step_ordinal,command_key
) VALUES
($1,1,'tenancy.activate_tenant'),
($1,2,'iam.initialize_tenant_identity'),
($1,3,'billing.assign_tenant_base_offering'),
($1,4,'billing.open_tenant_usage'),
($1,5,'systemresources.initialize_tenant_baseline'),
($1,6,'platformops.initialize_recovery_signing_identity'),
($1,7,'platformauthority.bind_owner'),
($1,8,'platformauthority.pin_protected_authority_head'),
($1,9,'activation.complete')`, plan.operationID)
	execute(`INSERT INTO platform.tenancy_onboarding_attempts(
attempt_id,operation_id,attempt_number,outcome,started_at
) VALUES($1,$2,1,'active',$3)`, plan.attemptID, plan.operationID, now)
	execute(`INSERT INTO platform.tenancy_onboarding_leases(
operation_id,attempt_id,lease_owner,lease_token_sha256,lease_until,
heartbeat_at,revision
) VALUES($1,$2,'capability-test',$3,$4::timestamptz+interval '10 minutes',$4,$5)`,
		plan.operationID, plan.attemptID, plan.leaseToken, now, plan.revision)
	execute(`INSERT INTO platform.tenancy_onboarding_receipts(
receipt_id,operation_id,attempt_id,receipt_kind,receipt_sha256,
evidence_sha256,occurred_at
) VALUES($1,$2,$3,'migration_completed',repeat('7',64),repeat('8',64),$4)`,
		plan.migrationReceiptID, plan.operationID, plan.attemptID, now)
	execute(`INSERT INTO platform.tenancy_onboarding_authority_acks(
operation_id,registry_binding_id,scope_kind,scope_id,generation,
state_sha256,proposal_sha256,ack_sha256,registry_sequence,verified_at
) VALUES($1,$2,'environment',$3,1,repeat('4',64),repeat('5',64),
repeat('6',64),1,$4)`, plan.operationID, plan.registryBindingID,
		plan.environmentID, now)
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit transaction capability plan: %v", err)
	}
}

func installCapabilityTestConsumers(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const consumer = `CREATE FUNCTION platform.test_consume_activation_v1(
p_operation_id uuid,p_manifest_sha256 text) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $body$
DECLARE v_operation platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
SELECT * INTO STRICT v_operation FROM platform.tenancy_onboarding_operations
WHERE operation_id=p_operation_id;
RETURN platform.tenancy_activation_command_consume_v1(
 v_operation.operation_id,'billing.assign_tenant_base_offering',
 v_operation.environment_id,v_operation.tenant_id,v_operation.operation_kind,
 v_operation.deployment_mode,v_operation.binding_purpose,p_manifest_sha256,
 v_operation.offering_version_id,v_operation.offering_revision,
 v_operation.offering_sha256,v_operation.offering_effective_at,
 v_operation.offering_source_sequence,v_operation.offering_source_sha256);
END;$body$;
REVOKE ALL ON FUNCTION platform.test_consume_activation_v1(uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.test_consume_activation_v1(uuid,text)
TO harpy_runtime;
CREATE TABLE platform.test_tenancy_authority_calls(
 operation_id uuid,call_key text,PRIMARY KEY(operation_id,call_key));`
	if _, err := database.ExecContext(ctx, consumer); err != nil {
		t.Fatalf("install capability test consumer: %v", err)
	}
	for _, stub := range []struct{ name, owner, key string }{
		{"billing_reconcile_current_global_authority_v2", "billing", "billing_global"},
		{"billing_reconcile_current_tenant_authorities_v2", "billing", "billing_tenant"},
		{"analytics_reconcile_current_tenant_authorities_v2", "analytics", "analytics"},
	} {
		installAuthorityCapabilityStub(t, ctx, database, stub.name, stub.owner, stub.key)
	}
	installActivationStepCapabilityTestConsumer(t, ctx, database)
}

func installAuthorityCapabilityStub(
	t *testing.T, ctx context.Context, database *sql.DB, name, owner, key string,
) {
	t.Helper()
	statement := fmt.Sprintf(`CREATE FUNCTION platform.%s()
RETURNS TABLE(reconciled_count integer,remaining_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $body$
DECLARE v_operation_id uuid;
BEGIN
IF platform.tenancy_authority_reconcile_scope_assert_v1('%s')
 IS DISTINCT FROM true THEN
 RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='test scope denied';
END IF;
SELECT operation_id INTO STRICT v_operation_id
FROM platform.tenancy_authority_reconcile_scopes
WHERE backend_pid=pg_catalog.pg_backend_pid()
AND transaction_id=pg_catalog.pg_current_xact_id() AND owner_key='%s';
INSERT INTO platform.test_tenancy_authority_calls VALUES(v_operation_id,'%s')
ON CONFLICT DO NOTHING;
RETURN QUERY SELECT 1,0;
END;$body$;
REVOKE ALL ON FUNCTION platform.%s() FROM PUBLIC;`, name, owner, owner, key, name)
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install %s: %v", name, err)
	}
}
