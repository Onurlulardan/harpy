-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_cutover_complete_v1_adoption_v1(
    p_operation_id uuid,p_environment_id uuid,p_frontend_sha256 text,p_mcp_sha256 text,
    p_legacy_boolean_hits bigint,p_module_migration_hits bigint,
    p_zero_hit_since timestamptz,p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_plan platform.platform_cutover_v1_adoption_plans%ROWTYPE;
    v_existing platform.platform_cutover_v1_adoption_completion_receipts%ROWTYPE;
    v_tenant_count integer;
    v_billing_count integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            !~ '^forgecrud-cutover/[0-9a-f-]{36}$'
       OR p_frontend_sha256 !~ '^[0-9a-f]{64}$'
       OR p_mcp_sha256 !~ '^[0-9a-f]{64}$'
       OR p_legacy_boolean_hits<>0 OR p_module_migration_hits<>0
       OR NOT pg_catalog.isfinite(p_zero_hit_since)
       OR NOT pg_catalog.isfinite(p_completed_at)
       OR p_zero_hit_since>p_completed_at-interval '30 days' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='V1 adoption completion denied';
    END IF;
    SELECT * INTO v_existing
      FROM platform.platform_cutover_v1_adoption_completion_receipts
     WHERE operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.frontend_contract_sha256 IS DISTINCT FROM p_frontend_sha256
           OR v_existing.mcp_contract_sha256 IS DISTINCT FROM p_mcp_sha256
           OR v_existing.zero_hit_since IS DISTINCT FROM p_zero_hit_since
           OR v_existing.completed_at IS DISTINCT FROM p_completed_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='V1 adoption completion replay conflicts';
        END IF;
        RETURN true;
    END IF;
    SELECT * INTO v_plan FROM platform.platform_cutover_v1_adoption_plans
     WHERE operation_id=p_operation_id AND environment_id=p_environment_id
       AND decision='ready' FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='ready V1 adoption plan is missing';
    END IF;
    SELECT receipt.adopted_tenant_count INTO v_tenant_count
      FROM platform.tenancy_v1_adoption_receipts AS receipt
     WHERE receipt.operation_id=p_operation_id AND receipt.environment_id=p_environment_id
       AND receipt.primary_tenant_id=v_plan.primary_tenant_id
       AND pg_catalog.encode(receipt.plan_sha256,'hex')=v_plan.plan_sha256 FOR SHARE;
    IF v_tenant_count IS NULL OR NOT EXISTS (
        SELECT 1 FROM platform.iam_v1_protected_authority_adoption_receipts AS receipt
         WHERE receipt.operation_id=p_operation_id AND receipt.environment_id=p_environment_id
           AND receipt.primary_tenant_id=v_plan.primary_tenant_id
           AND receipt.user_id=v_plan.owner_user_id
           AND pg_catalog.encode(receipt.plan_sha256,'hex')=v_plan.plan_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 tenancy or IAM adoption evidence is incomplete';
    END IF;
    SELECT pg_catalog.count(*)::integer INTO v_billing_count
      FROM platform.billing_v1_base_adoption_receipts AS receipt
     WHERE receipt.operation_id=p_operation_id AND receipt.environment_id=p_environment_id
       AND receipt.plan_sha256=v_plan.plan_sha256;
    IF v_billing_count IS DISTINCT FROM v_tenant_count THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 Billing adoption evidence is incomplete';
    END IF;
    INSERT INTO platform.platform_cutover_v1_adoption_completion_receipts(
        operation_id,environment_id,plan_sha256,frontend_contract_sha256,
        mcp_contract_sha256,legacy_boolean_hits,module_migration_hits,
        zero_hit_since,completed_at,recorded_by
    ) VALUES (p_operation_id,p_environment_id,v_plan.plan_sha256,p_frontend_sha256,
        p_mcp_sha256,0,0,p_zero_hit_since,p_completed_at,SESSION_USER);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_cutover_complete_v1_adoption_v1(
    uuid,uuid,text,text,bigint,bigint,timestamptz,timestamptz
) FROM PUBLIC;
