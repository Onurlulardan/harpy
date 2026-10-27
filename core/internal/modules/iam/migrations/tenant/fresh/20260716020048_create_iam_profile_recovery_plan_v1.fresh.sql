-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_plan_create_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_plan_document bytea, p_plan_sha256 bytea,
    p_actor_id uuid, p_manage_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, result_revision bigint, authorization_clock bigint,
    result_state text, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_plan jsonb;
    v_plan_id uuid;
    v_resource text;
    v_source_id uuid;
    v_target_id uuid;
    v_remaining_count bigint;
    v_revision bigint;
    v_state text;
    v_clock bigint;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR p_operation_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) <> 32 OR
       pg_catalog.octet_length(p_plan_sha256) <> 32 OR
       pg_catalog.octet_length(p_plan_document) NOT BETWEEN 2 AND 16777216 OR
       pg_catalog.sha256(p_plan_document) <> p_plan_sha256 OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.domain_cutovers AS cutover
           WHERE cutover.scope_kind = 'tenant' AND cutover.scope_id = p_tenant_id
             AND cutover.domain = 'iam' AND cutover.state = 'v2'
             AND cutover.writer = 'v2'
       ) OR NOT iam_profile_actor_assert_v1(
           p_scope_schema,p_actor_id,p_manage_permission_id,
           p_expected_token_version,p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile recovery plan authority is denied';
    END IF;
    v_plan := pg_catalog.convert_from(p_plan_document,'UTF8')::jsonb;
    v_plan_id := (v_plan->>'id')::uuid;
    IF v_plan_id IS NULL OR v_plan_id =
       '00000000-0000-0000-0000-000000000000'::uuid THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery plan identity is invalid';
    END IF;
    SELECT receipt.command_key,receipt.request_sha256,receipt.result_id,
           plan.revision,plan.state,receipt.clock_value
    INTO v_receipt_command,v_receipt_request,v_receipt_id,
         v_revision,v_state,v_clock
    FROM iam_authorization_operation_receipts AS receipt
    JOIN iam_profile_backfill_plans AS plan ON plan.id = receipt.result_id
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt_command <> 'iam.profile_backfill.recovery_create' OR
           v_receipt_request <> p_request_sha256 OR v_receipt_id <> v_plan_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P5742',
                MESSAGE = 'IAM profile recovery operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_plan_id,v_revision,v_clock,v_state,true;
        RETURN;
    END IF;
    BEGIN
        SELECT materialized.plan_id,materialized.resource_kind,
               materialized.source_schema_id,materialized.target_schema_id,
               materialized.remaining_source_count
        INTO v_plan_id,v_resource,v_source_id,v_target_id,v_remaining_count
        FROM iam_profile_recovery_materialize_v1(
            p_tenant_id,p_plan_document,p_plan_sha256,v_now
        ) AS materialized;
    EXCEPTION WHEN serialization_failure OR unique_violation THEN
        SELECT receipt.command_key,receipt.request_sha256,receipt.result_id,
               plan.revision,plan.state,receipt.clock_value
        INTO v_receipt_command,v_receipt_request,v_receipt_id,
             v_revision,v_state,v_clock
        FROM iam_authorization_operation_receipts AS receipt
        JOIN iam_profile_backfill_plans AS plan ON plan.id = receipt.result_id
        WHERE receipt.operation_id = p_operation_id;
        IF NOT FOUND OR v_receipt_command <>
           'iam.profile_backfill.recovery_create' OR
           v_receipt_request <> p_request_sha256 OR v_receipt_id <> v_plan_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P5742',
                MESSAGE = 'IAM profile recovery parent already has a child';
        END IF;
        RETURN QUERY SELECT v_plan_id,v_revision,v_clock,v_state,true;
        RETURN;
    END;
    UPDATE iam_authorization_clock SET clock_value = clock_value + 1,
        revision = revision + 1,updated_at = v_now WHERE singleton
        RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id,command_key,request_sha256,clock_value,result_code,
        result_id,result_revision,result_sha256,created_at
    ) VALUES (
        p_operation_id,'iam.profile_backfill.recovery_create',p_request_sha256,
        v_clock,'iam.profile_backfill.recovery_committed',v_plan_id,1,
        p_plan_sha256,v_now
    );
    v_revision := 1;
    v_state := 'planned';
    IF v_remaining_count = 0 THEN
        IF NOT iam_profile_backfill_activate_v2(
            v_plan_id,v_resource,v_source_id,v_target_id,0,v_now
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'IAM empty profile recovery activation failed';
        END IF;
        v_revision := 2;
        v_state := 'activated';
    END IF;
    RETURN QUERY SELECT v_plan_id,v_revision,v_clock,v_state,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_plan_create_v1(
    name,uuid,uuid,bytea,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
