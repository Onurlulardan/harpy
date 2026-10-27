-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_finalize_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage analytics_v1_adoption_stage%ROWTYPE;
    v_generation bigint;
    v_plan text;
    v_transform text;
    v_lease text;
    v_final_lsn pg_lsn;
    v_evidence jsonb;
    v_source_count bigint;
    v_source_root text;
    v_target_count bigint;
    v_target_root text;
    v_dependency text;
    v_public text;
    v_zero_hit text;
    v_rollback text;
    v_watermark text;
    v_old_root bytea;
    v_adopted bigint;
    v_retired bigint;
    v_acceptance text;
    v_now timestamptz;
BEGIN
    v_now:=pg_catalog.date_trunc(
      'microseconds',pg_catalog.transaction_timestamp());
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption finalization principal denied';
    END IF;
    SELECT ready.final_lsn,ready.evidence INTO STRICT v_final_lsn,v_evidence
    FROM analytics_v1_adoption_projection_ready_v1(
      p_operation_id,p_run_id,p_scope_id
    ) AS ready;
    SELECT operation.cutover_generation,operation.plan_sha256,
      operation.transform_manifest_sha256,operation.lease_manifest_sha256
    INTO STRICT v_generation,v_plan,v_transform,v_lease
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
    FOR SHARE;
    IF EXISTS (
      SELECT 1 FROM analytics_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='blocked'
    ) THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption has active blockers';
    END IF;
    SELECT root.row_count,root.root_sha256 INTO STRICT v_source_count,v_source_root
    FROM analytics_v1_adoption_source_root_v1(
      p_operation_id,p_run_id,p_scope_id
    ) AS root;
    SELECT pg_catalog.count(*),pg_catalog.encode(pg_catalog.sha256(
      COALESCE(pg_catalog.string_agg(
        pg_catalog.uuid_send(stage.legacy_endpoint_id)||stage.target_sha256,
        ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_target_count,v_target_root
    FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.disposition='adopted';
    v_dependency:=analytics_v1_adoption_dependency_sha256_v1(
      p_operation_id,p_run_id,p_scope_id);
    IF (v_evidence->>'source_count')::bigint IS DISTINCT FROM v_source_count OR
       v_evidence->>'source_root_sha256' IS DISTINCT FROM v_source_root OR
       (v_evidence->>'target_count')::bigint IS DISTINCT FROM v_target_count OR
       v_evidence->>'target_root_sha256' IS DISTINCT FROM v_target_root OR
       v_evidence->>'dependency_sha256' IS DISTINCT FROM v_dependency THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption terminal inventory differs';
    END IF;
    FOR v_stage IN
      SELECT stage.* FROM analytics_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='adopted'
      ORDER BY stage.legacy_endpoint_id
    LOOP
      PERFORM analytics_v1_adoption_project_query_v1(v_stage.target_document);
      PERFORM analytics_v1_adoption_project_dependencies_v1(v_stage.target_document);
      IF (v_stage.source_document->>'is_public')::boolean THEN
        PERFORM analytics_v1_adoption_project_release_v1(v_stage.target_document);
        PERFORM analytics_v1_adoption_project_public_v1(
          p_scope_id,v_stage.legacy_endpoint_id,
          v_stage.source_document,v_stage.target_document);
      END IF;
      PERFORM analytics_v1_adoption_project_differential_v1(
        p_operation_id,p_run_id,v_stage.legacy_endpoint_id,
        v_stage.source_sha256,v_stage.target_document->'differential');
    END LOOP;
    v_public:=analytics_v1_adoption_public_approval_sha256_v1(
      p_operation_id,p_run_id,p_scope_id);
    v_zero_hit:=analytics_v1_adoption_zero_hit_sha256_v1(
      p_operation_id,p_run_id,p_scope_id);
    v_rollback:=analytics_v1_adoption_rollback_sha256_v1(
      p_operation_id,p_run_id);
    v_watermark:=analytics_v1_adoption_watermark_sha256_v1(
      p_operation_id,p_run_id,v_final_lsn);
    IF v_evidence->>'public_approval_sha256' IS DISTINCT FROM v_public OR
       v_evidence->>'zero_hit_sha256' IS DISTINCT FROM v_zero_hit OR
       v_evidence->>'rollback_sha256' IS DISTINCT FROM v_rollback OR
       v_evidence->>'watermark_sha256' IS DISTINCT FROM v_watermark THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption acceptance evidence differs';
    END IF;
    INSERT INTO analytics_v1_adoption_items(
      run_id,legacy_endpoint_id,source_sha256,disposition,reason_code,
      query_id,query_version_id,publication_release_id,target_sha256,created_at
    ) SELECT p_run_id,stage.legacy_endpoint_id,stage.source_sha256,
      stage.disposition,stage.reason_code,
      CASE WHEN stage.disposition='adopted' THEN
        (stage.target_document->'query'->>'id')::uuid END,
      CASE WHEN stage.disposition='adopted' THEN
        (stage.target_document->'published_version'->>'id')::uuid END,
      CASE WHEN stage.disposition='adopted' AND
        (stage.source_document->>'is_public')::boolean THEN
        (stage.target_document->'release'->>'id')::uuid END,
      stage.target_sha256,v_now
    FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed ORDER BY stage.legacy_endpoint_id;
    SELECT pg_catalog.count(*) FILTER(WHERE item.disposition='adopted'),
      pg_catalog.count(*) FILTER(WHERE item.disposition='retired'),
      pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
        item.legacy_endpoint_id::text||'|'||pg_catalog.encode(item.source_sha256,'hex')||
        '|'||item.disposition||'|'||COALESCE(item.reason_code,'')||'|'||
        COALESCE(item.query_id::text,'')||'|'||COALESCE(item.query_version_id::text,'')||
        '|'||COALESCE(item.publication_release_id::text,'')||'|'||
        COALESCE(pg_catalog.encode(item.target_sha256,'hex'),''),
        E'\n' ORDER BY item.legacy_endpoint_id),''),'UTF8'))
    INTO v_adopted,v_retired,v_old_root
    FROM analytics_v1_adoption_items AS item WHERE item.run_id=p_run_id;
    INSERT INTO analytics_v1_adoption_acceptance_receipts(
      run_id,tenant_id,cutover_generation,source_count,adopted_count,
      retired_count,root_sha256,accepted_at
    ) VALUES (p_run_id,p_scope_id,v_generation,v_source_count,v_adopted,
      v_retired,v_old_root,v_now);
    v_acceptance:=pg_catalog.encode(pg_catalog.sha256(
      pg_catalog.convert_to('analytics-v1-adoption-acceptance.v1','UTF8')||
      pg_catalog.uuid_send(p_operation_id)||pg_catalog.uuid_send(p_run_id)||
      pg_catalog.uuid_send(p_scope_id)||pg_catalog.int8send(v_generation)||
      pg_catalog.convert_to(v_plan||v_transform||v_lease,'UTF8')||
      pg_catalog.pg_lsn_send(v_final_lsn)||
      pg_catalog.decode(v_evidence->>'evidence_sha256','hex')),'hex');
    INSERT INTO analytics_v1_adoption_receipts(
      operation_id,run_id,tenant_id,cutover_generation,plan_sha256,
      transform_sha256,lease_manifest_sha256,final_lsn,source_count,
      source_root_sha256,target_count,target_root_sha256,dependency_sha256,
      public_approval_sha256,zero_hit_sha256,rollback_sha256,watermark_sha256,
      evidence_sha256,acceptance_sha256,recorded_at
    ) VALUES (p_operation_id,p_run_id,p_scope_id,v_generation,v_plan,v_transform,
      v_lease,v_final_lsn,v_source_count,v_source_root,v_target_count,v_target_root,
      v_dependency,v_public,v_zero_hit,v_rollback,v_watermark,
      v_evidence->>'evidence_sha256',v_acceptance,v_now);
    RETURN v_target_root;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_finalize_v1(uuid,uuid,uuid)
FROM PUBLIC;
