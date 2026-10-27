-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_apply_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,
    p_snapshot_lsn pg_lsn,p_received_lsn pg_lsn,p_commit_lsn pg_lsn,
    p_source_xid xid8,p_batch_ordinal integer,p_change_batch jsonb,
    p_batch_sha256 text
) RETURNS TABLE(applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_change jsonb;
    v_payload jsonb;
    v_lease text;
    v_rows bigint;
    v_applied bigint;
    v_root text;
    v_now timestamptz;
BEGIN
    v_applied:=0;
    v_now:=pg_catalog.date_trunc(
      'microseconds',pg_catalog.transaction_timestamp());
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_kind IS DISTINCT FROM 'tenant' OR p_relation_ordinal<>1 OR
       p_checkpoint_key IS DISTINCT FROM 'openapi_endpoints' OR
       p_batch_sha256!~'^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics V1 adoption apply principal denied';
    END IF;
    SELECT operation.lease_manifest_sha256 INTO STRICT v_lease
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='analytics' AND operation.domain='analytics'
      AND operation.state IN ('copying','shadow','quiescing')
    FOR SHARE;
    FOR v_change IN
      SELECT item.value FROM pg_catalog.jsonb_array_elements(
        p_change_batch->'changes'
      ) WITH ORDINALITY AS item(value,position) ORDER BY item.position
    LOOP
      v_payload:=CASE v_change->>'kind'
        WHEN 'delete' THEN v_change->'before' ELSE v_change->'after' END;
      INSERT INTO analytics_v1_adoption_stage(
        operation_id,run_id,legacy_endpoint_id,source_sha256,source_document,
        disposition,reason_code,target_document,target_sha256,removed,
        source_xid,commit_lsn,batch_ordinal,change_ordinal,recorded_at
      ) VALUES (
        p_operation_id,p_run_id,
        (v_payload->'source_document'->>'id')::uuid,
        pg_catalog.decode(v_payload->>'source_sha256','hex'),
        v_payload->'source_document',v_payload->>'disposition',
        v_payload->>'reason_code',
        NULLIF(v_payload->'target_document','null'::jsonb),
        CASE WHEN v_payload->'target_sha256'='null'::jsonb THEN NULL
             ELSE pg_catalog.decode(v_payload->>'target_sha256','hex') END,
        v_change->>'kind'='delete',p_source_xid,p_commit_lsn,p_batch_ordinal,
        (v_change->>'ordinal')::smallint,v_now
      ) ON CONFLICT(operation_id,run_id,legacy_endpoint_id) DO UPDATE
      SET source_sha256=EXCLUDED.source_sha256,
        source_document=EXCLUDED.source_document,
        disposition=EXCLUDED.disposition,reason_code=EXCLUDED.reason_code,
        target_document=EXCLUDED.target_document,
        target_sha256=EXCLUDED.target_sha256,removed=EXCLUDED.removed,
        source_xid=EXCLUDED.source_xid,commit_lsn=EXCLUDED.commit_lsn,
        batch_ordinal=EXCLUDED.batch_ordinal,
        change_ordinal=EXCLUDED.change_ordinal,recorded_at=EXCLUDED.recorded_at
      WHERE (analytics_v1_adoption_stage.batch_ordinal,
             analytics_v1_adoption_stage.change_ordinal)
           <(EXCLUDED.batch_ordinal,EXCLUDED.change_ordinal);
      GET DIAGNOSTICS v_rows=ROW_COUNT;
      IF v_rows IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption endpoint checkpoint regressed';
      END IF;
      v_applied:=v_applied+1;
      INSERT INTO platform.applied_change_receipts(
        receipt_id,operation_id,run_id,manifest_sha256,source_xid,
        source_commit_lsn,relation_ordinal,change_ordinal,checkpoint_key,
        mutation_kind,mutation_sha256,applied_at
      ) VALUES (
        pg_catalog.gen_random_uuid(),p_operation_id,p_run_id,v_lease,
        p_source_xid,p_commit_lsn,1,
        (((p_batch_ordinal-1)::bigint*1000)+v_applied)::integer,
        'openapi_endpoints',v_change->>'kind',
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
          p_batch_sha256||':'||v_change::text,'UTF8'
        )),'hex'),v_now
      );
    END LOOP;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.uuid_send(stage.legacy_endpoint_id)||stage.target_sha256,
      ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_root FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.disposition='adopted';
    RETURN QUERY SELECT v_applied,v_root,v_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,
    xid8,integer,jsonb,text
) FROM PUBLIC;
