-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_authority_reconcile_current_v1(
    p_operation_id uuid,
    p_manifest_sha256 text
)
RETURNS TABLE (
    billing_global_reconciled integer,
    billing_global_remaining integer,
    billing_tenant_reconciled integer,
    billing_tenant_remaining integer,
    workflow_reconciled integer,
    workflow_remaining integer,
    analytics_reconciled integer,
    analytics_remaining integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_deleted integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenancy authority reconciliation is denied';
    END IF;
    SELECT * INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR SHARE;
    IF NOT FOUND
       OR v_operation.operation_kind IS DISTINCT FROM 'environment_initialize'
       OR v_operation.state IS DISTINCT FROM 'succeeded'
       OR v_operation.manifest_sha256 IS DISTINCT FROM p_manifest_sha256
       OR NOT EXISTS (
            SELECT 1
            FROM platform.tenancy_activation_facts AS fact
            JOIN platform.tenancy_tenant_bindings AS binding
              ON binding.environment_id = fact.environment_id
             AND binding.tenant_id = fact.tenant_id
             AND binding.purpose = fact.binding_purpose
            JOIN platform.tenants AS tenant ON tenant.id = fact.tenant_id
            WHERE fact.operation_id = v_operation.operation_id
              AND fact.environment_id = v_operation.environment_id
              AND fact.tenant_id = v_operation.tenant_id
              AND fact.owner_user_id = v_operation.first_user_id
              AND fact.binding_purpose = v_operation.binding_purpose
              AND fact.manifest_sha256 = v_operation.manifest_sha256
              AND tenant.deleted_at IS NULL
              AND tenant.status IN ('ACTIVE', 'TRIAL')
       )
       OR NOT EXISTS (
            SELECT 1 FROM platform.tenancy_onboarding_receipts AS receipt
            WHERE receipt.operation_id = v_operation.operation_id
              AND receipt.receipt_kind = 'activation_completed'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'completed environment initialization is required';
    END IF;
    INSERT INTO platform.tenancy_authority_reconcile_scopes (
        backend_pid, transaction_id, operation_id, environment_id,
        manifest_sha256, step_plan_sha256, owner_key, created_at
    )
    SELECT pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(),
           v_operation.operation_id, v_operation.environment_id,
           v_operation.manifest_sha256, v_operation.step_plan_sha256,
           owner_key, pg_catalog.transaction_timestamp()
    FROM (VALUES ('billing'), ('workflow'), ('analytics')) AS owners(owner_key);

    SELECT reconciled_count, remaining_count
    INTO billing_global_reconciled, billing_global_remaining
    FROM platform.billing_reconcile_current_global_authority_v2();
    SELECT reconciled_count, remaining_count
    INTO billing_tenant_reconciled, billing_tenant_remaining
    FROM platform.billing_reconcile_current_tenant_authorities_v2();
    SELECT reconciled_count, remaining_count
    INTO workflow_reconciled, workflow_remaining
    FROM platform.workflow_reconcile_current_tenant_authorities_v2();
    SELECT reconciled_count, remaining_count
    INTO analytics_reconciled, analytics_remaining
    FROM platform.analytics_reconcile_current_tenant_authorities_v2();

    DELETE FROM platform.tenancy_authority_reconcile_scopes AS scope
    WHERE scope.backend_pid = pg_catalog.pg_backend_pid()
      AND scope.transaction_id = pg_catalog.pg_current_xact_id()
      AND scope.operation_id = v_operation.operation_id;
    GET DIAGNOSTICS v_deleted = ROW_COUNT;
    IF v_deleted IS DISTINCT FROM 3 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'tenancy authority reconcile scope cleanup failed';
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_authority_reconcile_current_v1(
    uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_authority_reconcile_current_v1(
    uuid,text
) TO harpy_runtime;
