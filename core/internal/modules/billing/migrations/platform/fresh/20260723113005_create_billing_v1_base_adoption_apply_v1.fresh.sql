-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_base_adoption_apply_v1(
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
    v_assignment jsonb;
    v_plan_sha text;
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
            MESSAGE='V1 Billing base adoption apply denied';
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
            MESSAGE='V1 Billing base adoption document differs';
    END IF;
    v_plan := v_after->'plan';
    v_plan_sha := v_plan->>'plan_sha256';
    IF pg_catalog.jsonb_typeof(v_plan->'base_assignments')<>'array'
       OR pg_catalog.jsonb_array_length(v_plan->'base_assignments')<1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='V1 Billing base assignments are invalid';
    END IF;
    PERFORM 1 FROM platform.platform_cutover_v1_adoption_plans AS recorded
     WHERE recorded.operation_id=p_operation_id
       AND recorded.environment_id=p_scope_id AND recorded.decision='ready'
       AND recorded.plan_sha256=v_plan_sha
       AND recorded.plan_document=v_plan FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 Billing base adoption plan is not recorded';
    END IF;
    FOR v_assignment IN
        SELECT value FROM pg_catalog.jsonb_array_elements(
            v_plan->'base_assignments')
    LOOP
        PERFORM platform.billing_v1_base_adoption_assert_v1(
            p_operation_id,p_scope_id,(v_assignment->>'tenant_id')::uuid,
            v_assignment->>'kind',v_assignment#>>'{offering,key}',
            (v_assignment#>>'{offering,version_id}')::uuid,
            (v_assignment#>>'{offering,revision}')::bigint,
            v_assignment#>>'{offering,sha256}',
            (v_assignment#>>'{offering,effective_at}')::timestamptz,
            (v_assignment#>>'{offering,source_sequence}')::bigint,
            v_assignment#>>'{offering,source_sha256}',
            v_assignment->>'subscription_evidence_sha256',v_plan_sha);
    END LOOP;
    applied_count := pg_catalog.jsonb_array_length(
        v_plan->'base_assignments')::bigint;
    key_state_sha256 := v_plan_sha;
    state_sha256 := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'billing:'||v_plan_sha||':'||applied_count::text,'UTF8')),'hex');
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_base_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text
) FROM PUBLIC;
