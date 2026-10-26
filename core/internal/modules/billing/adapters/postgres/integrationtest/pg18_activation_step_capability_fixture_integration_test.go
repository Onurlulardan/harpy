//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func prepareBillingActivationStepCapabilityFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.test_tenancy_activation_steps (
 backend_pid integer NOT NULL,
 transaction_id xid8 NOT NULL,
 operation_id uuid NOT NULL,
 command_key text NOT NULL,
 PRIMARY KEY(backend_pid,transaction_id,operation_id,command_key)
);
CREATE FUNCTION platform.tenancy_activation_step_consume_v1(
 p_operation_id uuid,p_command_key text
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE v_consumed boolean;
BEGIN
 IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
    OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
  RETURN false;
 END IF;
 DELETE FROM platform.test_tenancy_activation_steps AS capability
 WHERE capability.backend_pid=pg_catalog.pg_backend_pid()
   AND capability.transaction_id=pg_catalog.pg_current_xact_id()
   AND capability.operation_id=p_operation_id
   AND capability.command_key=p_command_key
 RETURNING true INTO v_consumed;
 RETURN COALESCE(v_consumed,false);
END;
$function$;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_consume_v1(
 uuid,text
) FROM PUBLIC;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare Billing activation step capability: %v", err)
	}
}

func mintBillingActivationStep(
	ctx context.Context,
	transaction *sql.Tx,
	operationID uuid.UUID,
	commandKey string,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_activation_steps(
 backend_pid,transaction_id,operation_id,command_key
) VALUES(
 pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1,$2
)`, operationID, commandKey)
	return err
}

func createBillingUsageCapabilityFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) billingCapabilityFixture {
	t.Helper()
	tenantID := billingV7(t)
	insertBillingTenantTopology(
		t, ctx, database, tenantID, 8, billingAcceptanceA,
	)
	offering := loadBillingControlOffering(t, ctx, database)
	operationID := billingV7(t)
	authorityID := billingInstallationIDFixture(
		operationID, "tenant-authority",
	)
	assignmentID := billingInstallationIDFixture(
		operationID, "base-offering-assignment",
	)
	replayed, err := assignBillingInstallation(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering, true,
	)
	if err != nil || replayed {
		t.Fatalf("prepare Billing usage assignment=%t err=%v", replayed, err)
	}
	return billingCapabilityFixture{
		operationID: operationID, environmentID: uuid.MustParse(billingEnvironmentID),
		tenantID: tenantID, authorityID: authorityID, assignmentID: assignmentID,
		offering: offering, manifestSHA256: billingAcceptanceA,
	}
}
