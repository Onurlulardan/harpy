-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_apply_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_relation_ordinal smallint, p_checkpoint_key text,
    p_snapshot_lsn pg_lsn, p_received_lsn pg_lsn, p_commit_lsn pg_lsn,
    p_source_xid xid8, p_batch_ordinal integer,
    p_change_batch jsonb, p_batch_sha256 text
) RETURNS TABLE(applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_change jsonb;
    v_payload jsonb;
    v_lease_manifest text;
    v_rows bigint;
    v_applied bigint;
    v_key_sha text;
    v_state_sha text;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_kind IS DISTINCT FROM 'tenant' OR p_relation_ordinal<>1 OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption apply principal denied';
    END IF;
    v_now:=pg_catalog.date_trunc(
      'microseconds',pg_catalog.transaction_timestamp()
    );
    v_applied:=0;
    SELECT operation.lease_manifest_sha256 INTO STRICT v_lease_manifest
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind=p_scope_kind AND operation.scope_id=p_scope_id
      AND operation.group_id='catalog' AND operation.domain='catalog'
      AND operation.state IN ('copying','shadow','quiescing')
    FOR SHARE;
    FOR v_change IN
        SELECT value FROM pg_catalog.jsonb_array_elements(p_change_batch->'changes')
          WITH ORDINALITY AS item(value,position) ORDER BY position
    LOOP
        v_payload:=CASE v_change->>'kind'
          WHEN 'delete' THEN v_change->'before' ELSE v_change->'after' END;
        INSERT INTO catalog_v1_adoption_stage(
          operation_id,run_id,source_kind,source_identity,source_sha256,
          source_document,target_kind,target_identity,target_sha256,target_record,
          removed,source_xid,commit_lsn,batch_ordinal,change_ordinal,recorded_at
        ) VALUES (
          p_operation_id,p_run_id,v_payload->>'source_kind',
          v_payload->>'source_identity',
          pg_catalog.decode(v_payload->>'source_sha256','hex'),
          v_payload->'source_document',v_payload->>'target_kind',
          v_payload->>'target_identity',
          pg_catalog.sha256(pg_catalog.convert_to(
            (v_payload->'target_record')::text,'UTF8'
          )),v_payload->'target_record',v_change->>'kind'='delete',
          p_source_xid,p_commit_lsn,p_batch_ordinal,
          (v_change->>'ordinal')::smallint,v_now
        ) ON CONFLICT(operation_id,run_id,target_kind,target_identity) DO UPDATE
        SET source_kind=EXCLUDED.source_kind,
          source_identity=EXCLUDED.source_identity,
          source_sha256=EXCLUDED.source_sha256,
          source_document=EXCLUDED.source_document,
          target_sha256=EXCLUDED.target_sha256,
          target_record=EXCLUDED.target_record,removed=EXCLUDED.removed,
          source_xid=EXCLUDED.source_xid,commit_lsn=EXCLUDED.commit_lsn,
          batch_ordinal=EXCLUDED.batch_ordinal,
          change_ordinal=EXCLUDED.change_ordinal,recorded_at=EXCLUDED.recorded_at
        WHERE (catalog_v1_adoption_stage.batch_ordinal,
               catalog_v1_adoption_stage.change_ordinal)
             <(EXCLUDED.batch_ordinal,EXCLUDED.change_ordinal);
        GET DIAGNOSTICS v_rows=ROW_COUNT;
        IF v_rows IS DISTINCT FROM 1::bigint THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Catalog V1 adoption target checkpoint regressed';
        END IF;
        v_applied:=v_applied+1;
        INSERT INTO platform.applied_change_receipts(
          receipt_id,operation_id,run_id,manifest_sha256,source_xid,
          source_commit_lsn,relation_ordinal,change_ordinal,checkpoint_key,
          mutation_kind,mutation_sha256,applied_at
        ) VALUES (
          pg_catalog.gen_random_uuid(),p_operation_id,p_run_id,v_lease_manifest,
          p_source_xid,p_commit_lsn,p_relation_ordinal,
          (((p_batch_ordinal-1)::bigint*1000)+v_applied)::integer,
          p_checkpoint_key,v_change->>'kind',
          pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            p_batch_sha256||':'||v_change::text,'UTF8'
          )),'hex'),v_now
        );
    END LOOP;
    IF EXISTS (
      SELECT 1 FROM catalog_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed
      GROUP BY stage.source_kind,stage.source_identity
      HAVING pg_catalog.count(DISTINCT stage.source_sha256)<>1 OR
             pg_catalog.count(DISTINCT stage.source_document)<>1
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption source digest differs';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_kind))||
      pg_catalog.convert_to(stage.target_kind,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_identity))||
      pg_catalog.convert_to(stage.target_identity,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_record::text))||
      pg_catalog.convert_to(stage.target_record::text,'UTF8'),
      ''::bytea ORDER BY stage.target_kind,stage.target_identity),''::bytea)),'hex')
    INTO v_key_sha FROM catalog_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND stage.source_kind=p_checkpoint_key AND NOT stage.removed
      AND stage.target_kind<>'discard';
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_kind))||
      pg_catalog.convert_to(stage.target_kind,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_identity))||
      pg_catalog.convert_to(stage.target_identity,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(stage.target_record::text))||
      pg_catalog.convert_to(stage.target_record::text,'UTF8'),
      ''::bytea ORDER BY stage.target_kind,stage.target_identity),''::bytea)),'hex')
    INTO v_state_sha FROM catalog_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.target_kind<>'discard';
    RETURN QUERY SELECT v_applied,v_key_sha,v_state_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,
    xid8,integer,jsonb,text
) FROM PUBLIC;
