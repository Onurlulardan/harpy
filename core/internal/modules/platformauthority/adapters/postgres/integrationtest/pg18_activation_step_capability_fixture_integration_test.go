//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
)

func prepareAuthorityActivationStepFixture(
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
 environment_id uuid NOT NULL,
 tenant_id uuid NOT NULL,
 schema_name name NOT NULL,
 operation_kind text NOT NULL,
 PRIMARY KEY(backend_pid,transaction_id,operation_id,command_key)
);
CREATE FUNCTION platform.tenancy_activation_step_context_assert_v1(
 p_operation_id uuid,p_command_key text,p_environment_id uuid,
 p_tenant_id uuid,p_schema_name name,p_operation_kind text
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
 RETURN SESSION_USER = 'harpy_v2_runtime'
    AND CURRENT_USER = 'harpy_owner'
    AND EXISTS (
     SELECT 1 FROM platform.test_tenancy_activation_steps AS capability
     WHERE capability.backend_pid=pg_catalog.pg_backend_pid()
       AND capability.transaction_id=pg_catalog.pg_current_xact_id()
       AND capability.operation_id=p_operation_id
       AND capability.command_key=p_command_key
       AND capability.environment_id=p_environment_id
       AND capability.tenant_id=p_tenant_id
       AND capability.schema_name=p_schema_name
       AND capability.operation_kind=p_operation_kind
    );
END;
$function$;
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
CREATE FUNCTION platform.tenancy_activation_assert_v2(
 p_operation_id uuid,p_tenant_id uuid,p_schema_name name
) RETURNS boolean LANGUAGE sql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
 SELECT SESSION_USER = 'harpy_v2_runtime'
    AND CURRENT_USER = 'harpy_owner'
    AND p_operation_id IS NOT NULL
    AND p_tenant_id IS NOT NULL
    AND p_schema_name IS NOT NULL
$function$;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare authority activation capability: %v", err)
	}
}

func mintAuthorityActivationStep(
	ctx context.Context,
	transaction *sql.Tx,
	capability authorityActivationCapability,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_activation_steps(
 backend_pid,transaction_id,operation_id,command_key,
 environment_id,tenant_id,schema_name,operation_kind
) VALUES(
 pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1,$2,
 $3,$4,$5::name,$6
)`, capability.operationID, capability.commandKey,
		capability.environmentID, capability.tenantID,
		capability.schemaName, capability.operationKind)
	return err
}

type authorityActivationCapability struct {
	operationID, commandKey, environmentID string
	tenantID, schemaName, operationKind    string
}
