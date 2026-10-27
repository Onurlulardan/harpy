-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_v1_environment_plan_apply_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,p_commit_lsn pg_lsn,p_source_xid xid8,
    p_batch_ordinal integer,p_change_batch jsonb,p_batch_sha256 text
)
RETURNS TABLE (applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_after jsonb;
    v_plan jsonb;
    v_plan_sha text;
    v_environment_id uuid;
    v_primary_tenant_id uuid;
    v_owner_user_id uuid;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            IS DISTINCT FROM 'forgecrud-cutover/'||p_run_id::text
       OR p_scope_kind IS DISTINCT FROM 'global' OR p_scope_id IS NULL
       OR p_operation_id IS NULL OR p_run_id IS NULL
       OR p_relation_ordinal<1 OR p_checkpoint_key=''
       OR p_snapshot_lsn>p_commit_lsn OR p_received_lsn<p_commit_lsn
       OR p_batch_ordinal NOT BETWEEN 1 AND 2147483
       OR p_batch_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.jsonb_typeof(p_change_batch->'changes')<>'array'
       OR pg_catalog.jsonb_array_length(p_change_batch->'changes')<>1 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='V1 environment adoption apply denied';
    END IF;
    v_after := p_change_batch#>'{changes,0,after}';
    IF p_change_batch#>>'{changes,0,kind}' IS DISTINCT FROM 'insert'
       OR p_change_batch#>'{changes,0,before}' IS DISTINCT FROM 'null'::jsonb
       OR pg_catalog.jsonb_typeof(v_after)<>'object'
       OR (SELECT pg_catalog.array_agg(key ORDER BY key)
             FROM pg_catalog.jsonb_object_keys(v_after) AS key)
          IS DISTINCT FROM ARRAY['authority_ids','contract_version','inventory',
              'operation_id','plan']::text[]
       OR v_after->>'contract_version' IS DISTINCT FROM
            'platform-v1-environment-adoption.v1'
       OR (v_after->>'operation_id')::uuid IS DISTINCT FROM p_operation_id THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='V1 environment adoption document differs';
    END IF;
    v_plan := v_after->'plan';
    v_plan_sha := v_plan->>'plan_sha256';
    v_environment_id := (v_plan->>'environment_id')::uuid;
    v_primary_tenant_id := (v_plan->>'primary_tenant_id')::uuid;
    v_owner_user_id := (v_plan->>'owner_user_id')::uuid;
    IF v_environment_id IS DISTINCT FROM p_scope_id
       OR v_after#>>'{inventory,environment_id}' IS DISTINCT FROM p_scope_id::text
       OR v_after#>>'{inventory,mode}' IS DISTINCT FROM v_plan->>'mode'
       OR v_plan_sha !~ '^[0-9a-f]{64}$'
       OR v_plan->>'inventory_sha256' !~ '^[0-9a-f]{64}$'
       OR v_plan->>'owner_mapping_sha256' !~ '^[0-9a-f]{64}$'
       OR pg_catalog.jsonb_typeof(v_plan->'tenant_bindings')<>'array'
       OR pg_catalog.jsonb_typeof(v_plan->'base_assignments')<>'array'
       OR pg_catalog.jsonb_array_length(v_plan->'tenant_bindings')<1
       OR pg_catalog.jsonb_array_length(v_plan->'tenant_bindings')
            <>pg_catalog.jsonb_array_length(v_plan->'base_assignments') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='V1 environment adoption plan differs';
    END IF;
    PERFORM platform.platform_cutover_record_v1_adoption_plan_v1(
        p_operation_id,v_environment_id,v_plan->>'mode','ready',
        v_primary_tenant_id,v_owner_user_id,v_plan->>'inventory_sha256',
        v_plan_sha,'[]'::jsonb,v_plan,v_now);
    applied_count := 1;
    key_state_sha256 := v_plan_sha;
    state_sha256 := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        v_plan_sha||':'||applied_count::text,'UTF8')),'hex');
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_v1_environment_plan_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text
) FROM PUBLIC;
