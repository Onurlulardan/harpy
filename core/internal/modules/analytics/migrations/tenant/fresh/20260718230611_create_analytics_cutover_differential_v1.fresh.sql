-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_project_differential_v1(
    p_operation_id uuid,p_run_id uuid,p_legacy_endpoint_id uuid,
    p_source_sha256 bytea,p_differential jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_evidence bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_differential) AS key)
       IS DISTINCT FROM ARRAY[
        'compatible_image_sha256','evidence_sha256','reverse_source_sha256',
        'source_result_sha256','target_result_sha256','wire_contract_sha256'
       ]::text[] OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_each_text(p_differential) AS item
        WHERE item.value!~'^[0-9a-f]{64}$' OR item.value=repeat('0',64)
       ) OR p_differential->>'source_result_sha256' IS DISTINCT FROM
            p_differential->>'target_result_sha256' OR
       p_differential->>'reverse_source_sha256' IS DISTINCT FROM
         pg_catalog.encode(p_source_sha256,'hex') THEN
      RAISE EXCEPTION USING ERRCODE='22023',
        MESSAGE='Analytics V1 adoption differential evidence invalid';
    END IF;
    v_evidence:=pg_catalog.sha256(
      pg_catalog.convert_to('analytics-v1-adoption-differential.v1','UTF8')||
      pg_catalog.decode('00','hex')||pg_catalog.uuid_send(p_legacy_endpoint_id)||
      pg_catalog.decode('00','hex')||pg_catalog.decode(
        p_differential->>'source_result_sha256','hex')||
      pg_catalog.decode('00','hex')||pg_catalog.decode(
        p_differential->>'target_result_sha256','hex')||
      pg_catalog.decode('00','hex')||pg_catalog.decode(
        p_differential->>'wire_contract_sha256','hex')||
      pg_catalog.decode('00','hex')||pg_catalog.decode(
        p_differential->>'compatible_image_sha256','hex')||
      pg_catalog.decode('00','hex')||p_source_sha256
    );
    IF pg_catalog.encode(v_evidence,'hex') IS DISTINCT FROM
       p_differential->>'evidence_sha256' THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption differential digest differs';
    END IF;
    INSERT INTO analytics_v1_adoption_differentials(
      operation_id,run_id,legacy_endpoint_id,source_result_sha256,
      target_result_sha256,wire_contract_sha256,compatible_image_sha256,
      reverse_source_sha256,evidence_sha256,recorded_at
    ) VALUES (
      p_operation_id,p_run_id,p_legacy_endpoint_id,
      pg_catalog.decode(p_differential->>'source_result_sha256','hex'),
      pg_catalog.decode(p_differential->>'target_result_sha256','hex'),
      pg_catalog.decode(p_differential->>'wire_contract_sha256','hex'),
      pg_catalog.decode(p_differential->>'compatible_image_sha256','hex'),
      p_source_sha256,v_evidence,
      pg_catalog.date_trunc('microseconds',pg_catalog.transaction_timestamp())
    ) ON CONFLICT DO NOTHING;
    PERFORM 1 FROM analytics_v1_adoption_differentials AS differential
    WHERE differential.operation_id=p_operation_id AND differential.run_id=p_run_id
      AND differential.legacy_endpoint_id=p_legacy_endpoint_id
      AND differential.source_result_sha256=pg_catalog.decode(
        p_differential->>'source_result_sha256','hex')
      AND differential.target_result_sha256=differential.source_result_sha256
      AND differential.reverse_source_sha256=p_source_sha256
      AND differential.evidence_sha256=v_evidence;
    IF NOT FOUND THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption differential replay differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_project_differential_v1(
    uuid,uuid,uuid,bytea,jsonb
) FROM PUBLIC;
