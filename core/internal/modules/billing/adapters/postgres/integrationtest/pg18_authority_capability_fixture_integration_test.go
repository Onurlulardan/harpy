//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func prepareBillingAuthorityCapabilityFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const statement = `
CREATE TABLE platform.test_tenancy_activation_commands (
 backend_pid integer NOT NULL,
 transaction_id xid8 NOT NULL,
 operation_id uuid NOT NULL,
 command_name text NOT NULL,
 environment_id uuid NOT NULL,
 tenant_id uuid NOT NULL,
 operation_kind text NOT NULL,
 deployment_mode text NOT NULL,
 binding_purpose text NOT NULL,
 manifest_sha256 text NOT NULL,
 offering_version_id uuid NOT NULL,
 offering_revision bigint NOT NULL,
 offering_sha256 text NOT NULL,
 offering_effective_at timestamptz NOT NULL,
 source_sequence bigint,
 source_sha256 text,
 PRIMARY KEY(backend_pid,transaction_id,operation_id,command_name)
);
CREATE TABLE platform.test_tenancy_authority_reconcile_scopes (
 backend_pid integer NOT NULL,
 transaction_id xid8 NOT NULL,
 owner_key text NOT NULL,
 PRIMARY KEY(backend_pid,transaction_id,owner_key)
);
CREATE FUNCTION platform.tenancy_activation_command_consume_v1(
 p_operation_id uuid,p_command_name text,p_environment_id uuid,p_tenant_id uuid,
 p_operation_kind text,p_deployment_mode text,p_binding_purpose text,
 p_manifest_sha256 text,p_offering_version_id uuid,p_offering_revision bigint,
 p_offering_sha256 text,p_offering_effective_at timestamptz,
 p_source_sequence bigint,p_source_sha256 text
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_consumed boolean;
BEGIN
 IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
    OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
  RETURN false;
 END IF;
 DELETE FROM platform.test_tenancy_activation_commands AS command
 WHERE command.backend_pid=pg_catalog.pg_backend_pid()
   AND command.transaction_id=pg_catalog.pg_current_xact_id()
   AND ROW(command.operation_id,command.command_name,command.environment_id,
    command.tenant_id,command.operation_kind,command.deployment_mode,
    command.binding_purpose,command.manifest_sha256,
    command.offering_version_id,command.offering_revision,
    command.offering_sha256,command.offering_effective_at,
    command.source_sequence,command.source_sha256)
   IS NOT DISTINCT FROM ROW(p_operation_id,p_command_name,p_environment_id,
    p_tenant_id,p_operation_kind,p_deployment_mode,p_binding_purpose,
    p_manifest_sha256,p_offering_version_id,p_offering_revision,
    p_offering_sha256,p_offering_effective_at,p_source_sequence,p_source_sha256)
 RETURNING true INTO v_consumed;
 RETURN COALESCE(v_consumed,false);
END;
$function$;
REVOKE ALL ON FUNCTION platform.tenancy_activation_command_consume_v1(
 uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text
) FROM PUBLIC;
CREATE FUNCTION platform.tenancy_authority_reconcile_scope_assert_v1(
 p_owner_key text
) RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER
SET search_path TO pg_catalog
RETURN SESSION_USER='harpy_v2_runtime'
 AND CURRENT_USER='harpy_owner'
 AND EXISTS(
  SELECT 1 FROM platform.test_tenancy_authority_reconcile_scopes AS scope
  WHERE scope.backend_pid=pg_catalog.pg_backend_pid()
    AND scope.transaction_id=pg_catalog.pg_current_xact_id()
    AND scope.owner_key=p_owner_key
 );
REVOKE ALL ON FUNCTION
 platform.tenancy_authority_reconcile_scope_assert_v1(text) FROM PUBLIC;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare Billing authority capability fixture: %v", err)
	}
	prepareBillingActivationStepCapabilityFixture(t, ctx, database)
}

func mintBillingReconcileScope(
	ctx context.Context,
	transaction *sql.Tx,
	ownerKey string,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_authority_reconcile_scopes
 (backend_pid,transaction_id,owner_key)
VALUES(pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1)`,
		ownerKey)
	return err
}

func mintBillingInstallationCommand(
	ctx context.Context,
	transaction *sql.Tx,
	operationID, tenantID uuid.UUID,
	offering billingInstallationOffering,
	manifestSHA256 string,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_activation_commands(
 backend_pid,transaction_id,operation_id,command_name,environment_id,tenant_id,
 operation_kind,deployment_mode,binding_purpose,manifest_sha256,
 offering_version_id,offering_revision,offering_sha256,offering_effective_at,
 source_sequence,source_sha256
) VALUES(
 pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1,
 'billing.assign_tenant_base_offering',$2,$3,'environment_initialize',
 'saas','saas_control',$4,$5,$6,$7,$8,NULL,NULL
)`, operationID, uuid.MustParse(billingEnvironmentID), tenantID,
		manifestSHA256, offering.versionID, offering.revision,
		offering.sha256, offering.effective.UTC())
	return err
}

func billingInstallationIDFixture(
	operationID uuid.UUID,
	label string,
) uuid.UUID {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy:billing:installation:v1"))
	_, _ = hash.Write(operationID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(label))
	var result uuid.UUID
	copy(result[:], hash.Sum(nil))
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}
