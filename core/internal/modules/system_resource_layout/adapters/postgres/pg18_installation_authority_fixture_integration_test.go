//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func prepareLayoutInstallationCapabilityPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE TABLE platform.test_layout_capabilities (
    operation_id uuid NOT NULL,
    command_name text NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    schema_name name NOT NULL,
    operation_kind text NOT NULL,
    consumed boolean NOT NULL DEFAULT false,
    PRIMARY KEY (operation_id,command_name)
);
CREATE TABLE platform.tenancy_onboarding_operations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    operation_kind text NOT NULL
);
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY,
    schema_name name NOT NULL,
    status text NOT NULL,
    deleted_at timestamptz
);
CREATE TABLE platform.tenancy_activation_facts (
    operation_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    PRIMARY KEY (operation_id,tenant_id)
);
CREATE OR REPLACE FUNCTION platform.tenancy_activation_step_context_assert_v1(
    p_operation_id uuid,p_command text,p_environment_id uuid,p_tenant_id uuid,
    p_schema_name name,p_operation_kind text
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $$
SELECT EXISTS (
    SELECT 1 FROM platform.test_layout_capabilities AS capability
    WHERE capability.operation_id=p_operation_id
      AND capability.command_name=p_command
      AND capability.environment_id=p_environment_id
      AND capability.tenant_id=p_tenant_id
      AND capability.schema_name=p_schema_name
      AND capability.operation_kind=p_operation_kind
      AND NOT capability.consumed
)
$$;
CREATE OR REPLACE FUNCTION platform.tenancy_activation_step_consume_v1(
    p_operation_id uuid,p_command text
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $$
BEGIN
    UPDATE platform.test_layout_capabilities SET consumed=true
    WHERE operation_id=p_operation_id AND command_name=p_command
      AND NOT consumed;
    RETURN FOUND;
END
$$;
REVOKE ALL ON FUNCTION
 platform.tenancy_activation_step_context_assert_v1(
  uuid,text,uuid,uuid,name,text),
 platform.tenancy_activation_step_consume_v1(uuid,text)
FROM PUBLIC,harpy_runtime,harpy_v2_runtime;`)
	if err != nil {
		t.Fatalf("prepare layout installation capability: %v", err)
	}
}

func seedLayoutInstallationCapabilityPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO platform.tenancy_onboarding_operations
		 (operation_id,environment_id,operation_kind)
		 VALUES($1,$2,'environment_initialize')`,
			[]any{layoutContractOperationID, layoutContractEnvironmentID}},
		{`INSERT INTO platform.tenants(id,schema_name,status)
		 VALUES($1,'layout_contract','ACTIVE')`,
			[]any{layoutContractTenantID}},
		{`INSERT INTO platform.tenancy_activation_facts(operation_id,tenant_id)
		 VALUES($1,$2)`,
			[]any{layoutContractOperationID, layoutContractTenantID}},
		{`INSERT INTO platform.test_layout_capabilities(
		 operation_id,command_name,environment_id,tenant_id,
		 schema_name,operation_kind
		 ) VALUES(
		 $1,'systemresources.initialize_tenant_baseline',$2,$3,
		 'layout_contract','environment_initialize'
		 )`, []any{layoutContractOperationID, layoutContractEnvironmentID,
			layoutContractTenantID}},
	}
	for _, statement := range statements {
		_, err := database.ExecContext(ctx, statement.query, statement.args...)
		if err != nil {
			t.Fatalf("seed layout installation capability: %v", err)
		}
	}
}
