-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_cutover_record_v1_adoption_plan_v1(
    p_operation_id uuid,p_environment_id uuid,p_deployment_mode text,p_decision text,
    p_primary_tenant_id uuid,p_owner_user_id uuid,p_inventory_sha256 text,
    p_plan_sha256 text,p_issue_codes jsonb,p_plan_document jsonb,p_recorded_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platform_cutover_v1_adoption_plans%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            !~ '^forgecrud-cutover/[0-9a-f-]{36}$'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_deployment_mode NOT IN ('onpremise','saas')
       OR p_decision NOT IN ('ready','fenced')
       OR p_inventory_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.jsonb_typeof(p_issue_codes)<>'array'
       OR pg_catalog.jsonb_typeof(p_plan_document)<>'object'
       OR NOT pg_catalog.isfinite(p_recorded_at)
       OR (p_decision='ready' AND (p_primary_tenant_id IS NULL OR p_owner_user_id IS NULL
            OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
            OR pg_catalog.jsonb_array_length(p_issue_codes)<>0))
       OR (p_decision='fenced' AND pg_catalog.jsonb_array_length(p_issue_codes)=0) THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='V1 adoption plan denied';
    END IF;
    SELECT * INTO v_existing FROM platform.platform_cutover_v1_adoption_plans
     WHERE operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR v_existing.decision IS DISTINCT FROM p_decision
           OR v_existing.primary_tenant_id IS DISTINCT FROM p_primary_tenant_id
           OR v_existing.owner_user_id IS DISTINCT FROM p_owner_user_id
           OR v_existing.inventory_sha256 IS DISTINCT FROM p_inventory_sha256
           OR v_existing.plan_sha256 IS DISTINCT FROM p_plan_sha256
           OR v_existing.issue_codes IS DISTINCT FROM p_issue_codes
           OR v_existing.plan_document IS DISTINCT FROM p_plan_document
           OR v_existing.recorded_at IS DISTINCT FROM p_recorded_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='V1 adoption plan replay conflicts';
        END IF;
        RETURN true;
    END IF;
    INSERT INTO platform.platform_cutover_v1_adoption_plans(
        operation_id,environment_id,deployment_mode,decision,primary_tenant_id,
        owner_user_id,inventory_sha256,plan_sha256,issue_codes,plan_document,
        recorded_at,recorded_by
    ) VALUES (p_operation_id,p_environment_id,p_deployment_mode,p_decision,
        p_primary_tenant_id,p_owner_user_id,p_inventory_sha256,p_plan_sha256,
        p_issue_codes,p_plan_document,p_recorded_at,SESSION_USER);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_cutover_record_v1_adoption_plan_v1(
    uuid,uuid,text,text,uuid,uuid,text,text,jsonb,jsonb,timestamptz
) FROM PUBLIC;
