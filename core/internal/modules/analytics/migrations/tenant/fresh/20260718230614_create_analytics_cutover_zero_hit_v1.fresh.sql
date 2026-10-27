-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_zero_hit_sha256_v1(
    p_operation_id uuid,p_run_id uuid,p_tenant_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage analytics_v1_adoption_stage%ROWTYPE;
    v_snapshot jsonb;
    v_material bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption zero-hit evidence denied';
    END IF;
    v_material:=''::bytea;
    FOR v_stage IN
      SELECT stage.* FROM analytics_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='adopted'
        AND (stage.source_document->>'is_public')::boolean
      ORDER BY stage.legacy_endpoint_id
    LOOP
      v_snapshot:=platform.analytics_v1_adoption_public_bridge_v1(
        'inspect',p_tenant_id,v_stage.legacy_endpoint_id,NULL);
      IF (v_snapshot->>'route_exists')::boolean IS NOT TRUE OR
         (v_snapshot->>'publication_exists')::boolean IS NOT TRUE OR
         v_snapshot->>'route_state' IS DISTINCT FROM 'active' OR
         (v_snapshot->>'hit_count')::bigint IS DISTINCT FROM 0 OR
         v_snapshot->'last_hit_at' IS DISTINCT FROM 'null'::jsonb OR
         v_snapshot->'consumers_acked_at' IS NOT DISTINCT FROM 'null'::jsonb OR
         v_snapshot->>'consumer_inventory_sha256' IS NULL OR
         v_snapshot->>'consumer_inventory_sha256'=repeat('0',64) OR
         (v_snapshot->>'remove_after')::timestamptz<=
           (v_snapshot->>'consumers_acked_at')::timestamptz OR
         v_snapshot->>'legacy_tenant_identity' IS DISTINCT FROM p_tenant_id::text OR
         v_snapshot->>'method' IS DISTINCT FROM
           v_stage.source_document->>'method' OR
         v_snapshot->>'path' IS DISTINCT FROM
           v_stage.source_document->'config_json'->>'path' OR
         v_snapshot->>'slug' IS DISTINCT FROM v_stage.source_document->>'slug' OR
         v_snapshot->>'route_public_id' IS DISTINCT FROM
           v_stage.target_document->'public_pointer'->>'public_id' OR
         v_snapshot->>'consumer_inventory_sha256' IS DISTINCT FROM
           v_stage.target_document->'legacy_route'->>'consumer_inventory_sha256' OR
         v_snapshot->>'publication_public_id' IS DISTINCT FROM
           v_snapshot->>'route_public_id' OR
         v_snapshot->>'publication_state' IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption legacy route has traffic or no inventory';
      END IF;
      v_material:=v_material||pg_catalog.uuid_send(v_stage.legacy_endpoint_id)||
        pg_catalog.decode(v_snapshot->>'route_public_id','hex')||
        pg_catalog.decode(v_snapshot->>'consumer_inventory_sha256','hex')||
        pg_catalog.timestamptz_send(
          (v_snapshot->>'consumers_acked_at')::timestamptz)||
        pg_catalog.timestamptz_send((v_snapshot->>'remove_after')::timestamptz);
    END LOOP;
    RETURN pg_catalog.encode(pg_catalog.sha256(v_material),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_zero_hit_sha256_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
