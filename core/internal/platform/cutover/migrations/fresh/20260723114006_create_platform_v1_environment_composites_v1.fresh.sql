-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_v1_environment_apply_and_record_internal_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_snapshot_name text,
    p_snapshot_lsn pg_lsn,p_received_lsn pg_lsn,p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn,p_source_xid xid8,p_batch_ordinal integer,
    p_checkpoint_state text,p_last_key jsonb,p_change_batch_bytes bytea,
    p_batch_sha256 text,p_batch_bytes bigint,p_lease_manifest_sha256 text,
    p_expected_contract text
)
RETURNS TABLE (applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_batch jsonb;
    v_contract text;
    v_oid oid;
    v_signature text;
    v_replayed boolean;
    v_count bigint;
    v_key_sha text;
    v_state_sha text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_expected_contract NOT IN (
            'platform.v1_environment_plan.v1',
            'platform.v1_environment_completion.v1')
       OR p_batch_bytes IS DISTINCT FROM pg_catalog.octet_length(p_change_batch_bytes)
       OR p_batch_bytes NOT BETWEEN 1 AND 8388608
       OR pg_catalog.encode(pg_catalog.sha256(p_change_batch_bytes),'hex')
            IS DISTINCT FROM p_batch_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='V1 environment composite denied';
    END IF;
    v_batch := pg_catalog.convert_from(p_change_batch_bytes,'UTF8')::jsonb;
    IF (v_batch->>'operation_id')::uuid IS DISTINCT FROM p_operation_id
       OR (v_batch->>'run_id')::uuid IS DISTINCT FROM p_run_id
       OR v_batch->>'scope_kind' IS DISTINCT FROM p_scope_kind
       OR (v_batch->>'scope_id')::uuid IS DISTINCT FROM p_scope_id
       OR (v_batch->>'relation_ordinal')::smallint IS DISTINCT FROM p_relation_ordinal
       OR v_batch->>'checkpoint_key' IS DISTINCT FROM p_checkpoint_key
       OR v_batch->>'checkpoint_state' IS DISTINCT FROM p_checkpoint_state
       OR v_batch->'last_key' IS DISTINCT FROM p_last_key
       OR v_batch#>>'{source_position,snapshot_name}' IS DISTINCT FROM p_snapshot_name
       OR (v_batch#>>'{source_position,snapshot_lsn}')::pg_lsn IS DISTINCT FROM p_snapshot_lsn
       OR (v_batch#>>'{source_position,received_lsn}')::pg_lsn IS DISTINCT FROM p_received_lsn
       OR (v_batch#>>'{source_position,commit_lsn}')::pg_lsn IS DISTINCT FROM p_commit_lsn
       OR (v_batch#>>'{source_position,final_lsn}')::pg_lsn IS DISTINCT FROM p_final_lsn
       OR (v_batch#>>'{source_position,source_xid}')::xid8 IS DISTINCT FROM p_source_xid
       OR (v_batch#>>'{source_position,batch_ordinal}')::integer IS DISTINCT FROM p_batch_ordinal THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='V1 environment batch evidence differs';
    END IF;
    SELECT prepared.procedure_contract_id,prepared.procedure_oid,
        prepared.procedure_signature,prepared.already_applied,
        prepared.applied_count,prepared.key_state_sha256,prepared.state_sha256
      INTO STRICT v_contract,v_oid,v_signature,v_replayed,
        v_count,v_key_sha,v_state_sha
      FROM platform.cutover_prepare_batch_v1(
        p_operation_id,p_run_id,p_scope_kind,p_scope_id,p_relation_ordinal,
        p_checkpoint_key,p_batch_ordinal,p_commit_lsn,p_source_xid,
        p_batch_sha256,p_lease_manifest_sha256) AS prepared;
    IF v_contract IS DISTINCT FROM p_expected_contract
       OR (p_expected_contract='platform.v1_environment_plan.v1' AND (
            v_oid IS DISTINCT FROM 'platform.platform_v1_environment_plan_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid
            OR v_signature IS DISTINCT FROM 'platform.platform_v1_environment_plan_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'))
       OR (p_expected_contract='platform.v1_environment_completion.v1' AND (
            v_oid IS DISTINCT FROM 'platform.platform_v1_environment_completion_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid
            OR v_signature IS DISTINCT FROM 'platform.platform_v1_environment_completion_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)')) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 environment procedure binding differs';
    END IF;
    IF v_replayed THEN
        RETURN QUERY SELECT v_count,v_key_sha,v_state_sha;
        RETURN;
    END IF;
    IF p_expected_contract='platform.v1_environment_plan.v1' THEN
        SELECT result.applied_count,result.key_state_sha256,result.state_sha256
          INTO STRICT v_count,v_key_sha,v_state_sha
          FROM platform.platform_v1_environment_plan_apply_v1(
            p_operation_id,p_run_id,p_scope_kind,p_scope_id,p_relation_ordinal,
            p_checkpoint_key,p_snapshot_lsn,p_received_lsn,p_commit_lsn,
            p_source_xid,p_batch_ordinal,v_batch,p_batch_sha256) AS result;
    ELSE
        SELECT result.applied_count,result.key_state_sha256,result.state_sha256
          INTO STRICT v_count,v_key_sha,v_state_sha
          FROM platform.platform_v1_environment_completion_apply_v1(
            p_operation_id,p_run_id,p_scope_kind,p_scope_id,p_relation_ordinal,
            p_checkpoint_key,p_snapshot_lsn,p_received_lsn,p_commit_lsn,
            p_source_xid,p_batch_ordinal,v_batch,p_batch_sha256) AS result;
    END IF;
    IF v_count<0 OR v_key_sha !~ '^[0-9a-f]{64}$'
       OR v_state_sha !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22000',
            MESSAGE='V1 environment apply result is invalid';
    END IF;
    PERFORM platform.cutover_record_batch_v1(
        p_operation_id,p_run_id,p_relation_ordinal,p_checkpoint_key,p_snapshot_name,
        p_snapshot_lsn,p_received_lsn,p_commit_lsn,p_final_lsn,p_source_xid,
        p_batch_ordinal,p_checkpoint_state,p_last_key,p_batch_sha256,p_batch_bytes,
        v_count,v_key_sha,v_state_sha);
    RETURN QUERY SELECT v_count,v_key_sha,v_state_sha;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_v1_environment_apply_and_record_internal_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,
    integer,text,jsonb,bytea,text,bigint,text,text
) FROM PUBLIC;

CREATE FUNCTION platform.platform_v1_environment_plan_apply_and_record_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_snapshot_name text,
    p_snapshot_lsn pg_lsn,p_received_lsn pg_lsn,p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn,p_source_xid xid8,p_batch_ordinal integer,
    p_checkpoint_state text,p_last_key jsonb,p_change_batch_bytes bytea,
    p_batch_sha256 text,p_batch_bytes bigint,p_lease_manifest_sha256 text
)
RETURNS TABLE (applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$ BEGIN
    RETURN QUERY SELECT * FROM platform.platform_v1_environment_apply_and_record_internal_v1(
        p_operation_id,p_run_id,p_scope_kind,p_scope_id,p_relation_ordinal,
        p_checkpoint_key,p_snapshot_name,p_snapshot_lsn,p_received_lsn,p_commit_lsn,
        p_final_lsn,p_source_xid,p_batch_ordinal,p_checkpoint_state,p_last_key,
        p_change_batch_bytes,p_batch_sha256,p_batch_bytes,p_lease_manifest_sha256,
        'platform.v1_environment_plan.v1');
END; $function$;

CREATE FUNCTION platform.platform_v1_environment_completion_apply_and_record_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_snapshot_name text,
    p_snapshot_lsn pg_lsn,p_received_lsn pg_lsn,p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn,p_source_xid xid8,p_batch_ordinal integer,
    p_checkpoint_state text,p_last_key jsonb,p_change_batch_bytes bytea,
    p_batch_sha256 text,p_batch_bytes bigint,p_lease_manifest_sha256 text
)
RETURNS TABLE (applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$ BEGIN
    RETURN QUERY SELECT * FROM platform.platform_v1_environment_apply_and_record_internal_v1(
        p_operation_id,p_run_id,p_scope_kind,p_scope_id,p_relation_ordinal,
        p_checkpoint_key,p_snapshot_name,p_snapshot_lsn,p_received_lsn,p_commit_lsn,
        p_final_lsn,p_source_xid,p_batch_ordinal,p_checkpoint_state,p_last_key,
        p_change_batch_bytes,p_batch_sha256,p_batch_bytes,p_lease_manifest_sha256,
        'platform.v1_environment_completion.v1');
END; $function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_v1_environment_plan_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,
    integer,text,jsonb,bytea,text,bigint,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_v1_environment_completion_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,
    integer,text,jsonb,bytea,text,bigint,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platform_cutover_record_v1_adoption_plan_v1(
    uuid,uuid,text,text,uuid,uuid,text,text,jsonb,jsonb,timestamptz
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION platform.platform_cutover_complete_v1_adoption_v1(
    uuid,uuid,text,text,bigint,bigint,timestamptz,timestamptz
) FROM forgecrud_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_v1_environment_plan_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,
    integer,text,jsonb,bytea,text,bigint,text
) TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION platform.platform_v1_environment_completion_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,pg_lsn,xid8,
    integer,text,jsonb,bytea,text,bigint,text
) TO forgecrud_adopter;
