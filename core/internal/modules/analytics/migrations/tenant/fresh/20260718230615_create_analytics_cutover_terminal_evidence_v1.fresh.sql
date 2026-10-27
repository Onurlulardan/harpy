-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_source_root_v1(
    p_operation_id uuid,p_run_id uuid,p_tenant_id uuid
) RETURNS TABLE(row_count bigint,root_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage_count bigint;
    v_stage_root text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption source inventory denied';
    END IF;
    SELECT pg_catalog.count(*),pg_catalog.encode(pg_catalog.sha256(
      COALESCE(pg_catalog.string_agg(
        pg_catalog.uuid_send(stage.legacy_endpoint_id)||stage.source_sha256,
        ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_stage_count,v_stage_root
    FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed;
    RETURN QUERY
    WITH live AS (
      SELECT endpoint.id,pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.to_jsonb(endpoint)::text,'UTF8')) AS source_sha256
      FROM openapi_endpoints AS endpoint WHERE endpoint.tenant_id=p_tenant_id
    ), root AS (
      SELECT pg_catalog.count(*) AS count,
        pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
          pg_catalog.uuid_send(live.id)||live.source_sha256,
          ''::bytea ORDER BY live.id),''::bytea)),'hex') AS digest
      FROM live
    ) SELECT root.count,root.digest FROM root
      WHERE root.count=v_stage_count AND root.digest=v_stage_root;
    IF NOT FOUND THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption live source inventory differs';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_v1_adoption_rollback_sha256_v1(
    p_operation_id uuid,p_run_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR EXISTS (
      SELECT 1 FROM analytics_v1_adoption_stage AS stage
      LEFT JOIN analytics_v1_adoption_differentials AS differential
        ON differential.operation_id=stage.operation_id
       AND differential.run_id=stage.run_id
       AND differential.legacy_endpoint_id=stage.legacy_endpoint_id
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='adopted' AND (
          differential.legacy_endpoint_id IS NULL OR
          differential.source_result_sha256<>differential.target_result_sha256 OR
          differential.reverse_source_sha256<>stage.source_sha256 OR
          differential.evidence_sha256<>pg_catalog.sha256(
            pg_catalog.convert_to('analytics-v1-adoption-differential.v1','UTF8')||
            pg_catalog.decode('00','hex')||pg_catalog.uuid_send(stage.legacy_endpoint_id)||
            pg_catalog.decode('00','hex')||differential.source_result_sha256||
            pg_catalog.decode('00','hex')||differential.target_result_sha256||
            pg_catalog.decode('00','hex')||differential.wire_contract_sha256||
            pg_catalog.decode('00','hex')||differential.compatible_image_sha256||
            pg_catalog.decode('00','hex')||stage.source_sha256)
        )
    ) OR (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_differentials
          WHERE operation_id=p_operation_id AND run_id=p_run_id)<>
         (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_stage
          WHERE operation_id=p_operation_id AND run_id=p_run_id
            AND NOT removed AND disposition='adopted') THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption rollback evidence differs';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.uuid_send(differential.legacy_endpoint_id)||
      differential.source_result_sha256||differential.target_result_sha256||
      differential.wire_contract_sha256||differential.compatible_image_sha256||
      differential.reverse_source_sha256||differential.evidence_sha256,
      ''::bytea ORDER BY differential.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_digest FROM analytics_v1_adoption_differentials AS differential
    WHERE differential.operation_id=p_operation_id AND differential.run_id=p_run_id;
    RETURN v_digest;
END;
$function$;

CREATE FUNCTION analytics_v1_adoption_watermark_sha256_v1(
    p_operation_id uuid,p_run_id uuid,p_final_lsn pg_lsn
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_final_lsn IS NULL OR EXISTS (
        SELECT 1 FROM analytics_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
          AND stage.commit_lsn>p_final_lsn
       ) OR (SELECT pg_catalog.count(*) FROM platform.applied_change_receipts
             WHERE operation_id=p_operation_id AND run_id=p_run_id
               AND relation_ordinal=1 AND checkpoint_key='openapi_endpoints')<
            (SELECT pg_catalog.count(*) FROM analytics_v1_adoption_stage
             WHERE operation_id=p_operation_id AND run_id=p_run_id) THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption watermark evidence differs';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(
      pg_catalog.convert_to('analytics-v1-adoption-watermark.v1','UTF8')||
      pg_catalog.pg_lsn_send(p_final_lsn)||COALESCE(pg_catalog.string_agg(
        pg_catalog.uuid_send(stage.legacy_endpoint_id)||
        pg_catalog.xid8send(stage.source_xid)||pg_catalog.pg_lsn_send(stage.commit_lsn)||
        pg_catalog.int4send(stage.batch_ordinal)||
        pg_catalog.int2send(stage.change_ordinal),
        ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_digest FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id;
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_source_root_v1(uuid,uuid,uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_v1_adoption_rollback_sha256_v1(uuid,uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_v1_adoption_watermark_sha256_v1(
    uuid,uuid,pg_lsn
) FROM PUBLIC;
