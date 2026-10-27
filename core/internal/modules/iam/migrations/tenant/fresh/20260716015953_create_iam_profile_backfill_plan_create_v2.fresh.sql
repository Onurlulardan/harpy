-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_plan_create_v2(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_plan_document bytea, p_plan_sha256 bytea,
    p_actor_id uuid, p_manage_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (result_id uuid, result_revision bigint, replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_plan jsonb;
    v_plan_id uuid;
    v_resource_kind text;
    v_source_schema_id uuid;
    v_target_schema_id uuid;
    v_target_version bigint;
    v_target_definition_sha256 bytea;
    v_source_count bigint;
    v_page_size integer;
    v_page_count integer;
    v_expected_source bytea;
    v_expected_target bytea;
    v_initial_rolling bytea;
    v_clock bigint;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_draft_revision bigint;
    v_mapping_count integer;
BEGIN
    IF p_tenant_id IS NULL OR p_operation_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256)<>32 OR
       pg_catalog.octet_length(p_plan_sha256)<>32 OR
       pg_catalog.octet_length(p_plan_document) NOT BETWEEN 2 AND 16777216 OR
       pg_catalog.sha256(p_plan_document)<>p_plan_sha256 OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id=p_tenant_id AND tenant.schema_name=p_scope_schema
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.domain_cutovers AS cutover
           WHERE cutover.scope_kind='tenant' AND cutover.scope_id=p_tenant_id
             AND cutover.domain='iam' AND cutover.state='v2' AND cutover.writer='v2'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM profile backfill plan input is invalid';
    END IF;
    IF NOT iam_profile_actor_assert_v1(
        p_scope_schema,p_actor_id,p_manage_permission_id,
        p_expected_token_version,p_expected_authorization_version
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM profile backfill plan authority is denied';
    END IF;
    v_plan:=pg_catalog.convert_from(p_plan_document,'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_plan)<>'object' OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_object_keys(v_plan) AS key
        WHERE key NOT IN (
            'id','resource_kind','source_schema_id','target_schema_id','target_version',
            'target_definition','target_definition_sha256','converter_version',
            'source_count','expected_source_sha256','expected_target_sha256',
            'page_size','initial_rolling_sha256','mappings','pages'
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM profile backfill plan document is invalid';
    END IF;
    v_plan_id:=(v_plan->>'id')::uuid;
    v_resource_kind:=v_plan->>'resource_kind';
    v_source_schema_id:=(v_plan->>'source_schema_id')::uuid;
    v_target_schema_id:=(v_plan->>'target_schema_id')::uuid;
    v_target_version:=(v_plan->>'target_version')::bigint;
    v_target_definition_sha256:=pg_catalog.decode(
        v_plan->>'target_definition_sha256','hex');
    v_source_count:=(v_plan->>'source_count')::bigint;
    v_page_size:=(v_plan->>'page_size')::integer;
    v_expected_source:=pg_catalog.decode(v_plan->>'expected_source_sha256','hex');
    v_expected_target:=pg_catalog.decode(v_plan->>'expected_target_sha256','hex');
    v_initial_rolling:=pg_catalog.decode(v_plan->>'initial_rolling_sha256','hex');
    v_page_count:=pg_catalog.jsonb_array_length(v_plan->'pages');
    IF v_plan_id IS NULL OR v_resource_kind NOT IN ('user','organization','role') OR
       v_source_schema_id IS NULL OR v_target_schema_id IS NULL OR
       v_source_schema_id=v_target_schema_id OR v_target_version<2 OR
       v_plan->>'converter_version'<>'typed_map_v1' OR v_source_count<0 OR
       v_page_size NOT BETWEEN 1 AND 250 OR v_page_count NOT BETWEEN 1 AND 65536 OR
       pg_catalog.jsonb_typeof(v_plan->'target_definition')<>'object' OR
       pg_catalog.jsonb_typeof(v_plan->'mappings')<>'array' OR
       pg_catalog.octet_length(v_target_definition_sha256)<>32 OR
       pg_catalog.octet_length(v_expected_source)<>32 OR
       pg_catalog.octet_length(v_expected_target)<>32 OR
       pg_catalog.octet_length(v_initial_rolling)<>32 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM profile backfill plan contract is invalid';
    END IF;
    SELECT receipt.command_key,receipt.request_sha256,receipt.result_id,
           receipt.result_revision INTO v_receipt_command,v_receipt_request,
           v_receipt_id,v_receipt_revision
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id;
    IF FOUND THEN
        IF v_receipt_command<>'iam.profile_backfill.plan_create' OR
           v_receipt_request<>p_request_sha256 OR v_receipt_id<>v_plan_id THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='IAM profile backfill plan identity was reused';
        END IF;
        RETURN QUERY SELECT v_plan_id,v_receipt_revision,true;
        RETURN;
    END IF;
    SELECT iam_profile_backfill_draft_create_v2(
        p_scope_schema,v_target_schema_id,v_resource_kind,v_target_version,
        v_plan->'target_definition',v_target_definition_sha256,v_source_schema_id
    ) INTO v_draft_revision;
    INSERT INTO iam_profile_backfill_plans (
        id,resource_kind,source_schema_id,target_schema_id,converter_version,
        plan_sha256,source_count,expected_source_sha256,expected_target_sha256,
        page_size,page_count,state,activated_at,fenced_at,revision,created_at,updated_at
    ) VALUES (
        v_plan_id,v_resource_kind,v_source_schema_id,v_target_schema_id,'typed_map_v1',
        p_plan_sha256,v_source_count,v_expected_source,v_expected_target,
        v_page_size,v_page_count,'planned',NULL,NULL,1,v_now,v_now
    );
    SELECT iam_profile_backfill_mapping_create_v1(
        v_plan_id,v_resource_kind,v_source_schema_id,v_target_schema_id,
        v_plan->'mappings'
    ) INTO v_mapping_count;
    IF v_draft_revision<>1 OR v_mapping_count<1 THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='IAM profile backfill materialization is inconsistent';
    END IF;
    IF iam_profile_backfill_page_create_v2(
        p_tenant_id,v_plan_id,v_resource_kind,v_source_schema_id,v_source_count,
        v_page_size,v_expected_source,v_expected_target,v_initial_rolling,
        v_plan->'pages',v_now
    )<>v_page_count THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='IAM profile backfill page count is inconsistent';
    END IF;
    INSERT INTO iam_profile_backfill_checkpoints (
        plan_id,completed_pages,converted_count,error_count,rolling_sha256,
        last_page_number,revision,updated_at
    ) VALUES (v_plan_id,0,0,0,v_initial_rolling,NULL,1,v_now);
    UPDATE iam_authorization_clock SET clock_value=clock_value+1,
        revision=revision+1,updated_at=v_now WHERE singleton
        RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM authorization clock is unavailable';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id,command_key,request_sha256,clock_value,result_code,
        result_id,result_revision,result_sha256,created_at
    ) VALUES (
        p_operation_id,'iam.profile_backfill.plan_create',p_request_sha256,v_clock,
        'iam.profile_backfill.planned',v_plan_id,1,p_plan_sha256,v_now
    );
    RETURN QUERY SELECT v_plan_id,1::bigint,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_plan_create_v2(
    name,uuid,uuid,bytea,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;
