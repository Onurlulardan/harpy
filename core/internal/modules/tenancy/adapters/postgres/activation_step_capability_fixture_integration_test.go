//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

var activationStepCommands = []string{
	"iam.initialize_tenant_identity",
	"billing.open_tenant_usage",
	"systemresources.initialize_tenant_baseline",
	"platformops.initialize_recovery_signing_identity",
	"platformauthority.bind_owner",
	"platformauthority.pin_protected_authority_head",
	"activation.complete",
}

func completeTransactionCapabilityOperation(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID uuid.UUID, revision int64,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	_, err := transaction.ExecContext(ctx, `
SELECT platform.tenancy_onboarding_complete_v1(
 $1,$2,repeat('a',64),repeat('b',64),$3
)`, operationID, uuid.New(), revision)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("complete capability activation: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit capability activation: %v", err)
	}
}

func installActivationStepCapabilityTestConsumer(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `CREATE FUNCTION platform.test_consume_step_v1(
p_operation_id uuid,p_command_key text) RETURNS boolean
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $body$
SELECT platform.tenancy_activation_step_consume_v1($1,$2)
$body$;
REVOKE ALL ON FUNCTION platform.test_consume_step_v1(uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.test_consume_step_v1(uuid,text)
TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install activation step capability consumer: %v", err)
	}
}
