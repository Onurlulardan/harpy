-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_project_public_v1(
    p_tenant_id uuid,p_legacy_endpoint_id uuid,p_source jsonb,p_target jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_pointer jsonb;
    v_route jsonb;
    v_snapshot jsonb;
BEGIN
    v_pointer:=p_target->'public_pointer';
    v_route:=p_target->'legacy_route';
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_pointer) AS key)
       IS DISTINCT FROM ARRAY[
        'activated_at','audience','capability_expires_at','capability_sha256',
        'generation','public_id','release_id','resource_kind',
        'resource_version_id','revoked_at','state','tenant_id'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_route) AS key)
       IS DISTINCT FROM ARRAY[
        'consumer_inventory_sha256','consumers_acked_at','hit_count','last_hit_at',
        'legacy_endpoint_id','legacy_tenant_identity','method','path','public_id',
        'release_id','remove_after','retired_at','retirement_evidence_sha256',
        'retirement_observed_at','slug','state','tenant_id'
       ]::text[] OR
       (v_pointer->>'tenant_id')::uuid IS DISTINCT FROM p_tenant_id OR
       v_pointer->>'resource_kind'<>'query' OR v_pointer->>'audience'<>'public' OR
       v_pointer->>'state'<>'active' OR
       v_pointer->'capability_sha256'<>'null'::jsonb OR
       v_pointer->'capability_expires_at'<>'null'::jsonb OR
       v_pointer->'revoked_at'<>'null'::jsonb OR
       v_pointer->>'public_id'!~'^[0-9a-f]{32}$' OR
       (v_pointer->>'release_id')::uuid IS DISTINCT FROM
         (p_target->'release'->>'id')::uuid OR
       (v_pointer->>'resource_version_id')::uuid IS DISTINCT FROM
         (p_target->'published_version'->>'id')::uuid OR
       (v_route->>'tenant_id')::uuid IS DISTINCT FROM p_tenant_id OR
       (v_route->>'legacy_endpoint_id')::uuid IS DISTINCT FROM p_legacy_endpoint_id OR
       v_route->>'legacy_tenant_identity' IS DISTINCT FROM p_tenant_id::text OR
       v_route->>'method' IS DISTINCT FROM p_source->>'method' OR
       v_route->>'slug' IS DISTINCT FROM p_source->>'slug' OR
       v_route->>'path' IS DISTINCT FROM p_source->'config_json'->>'path' OR
       v_route->>'public_id' IS DISTINCT FROM v_pointer->>'public_id' OR
       (v_route->>'release_id')::uuid IS DISTINCT FROM
         (v_pointer->>'release_id')::uuid OR
       v_route->>'state'<>'active' OR (v_route->>'hit_count')::bigint<>0 OR
       v_route->'last_hit_at'<>'null'::jsonb OR
       v_route->'consumers_acked_at'='null'::jsonb OR
       v_route->>'consumer_inventory_sha256'!~'^[0-9a-f]{64}$' OR
       v_route->'retirement_evidence_sha256'<>'null'::jsonb OR
       v_route->'retirement_observed_at'<>'null'::jsonb OR
       v_route->'retired_at'<>'null'::jsonb THEN
      RAISE EXCEPTION USING ERRCODE='22023',
        MESSAGE='Analytics V1 adoption public projection invalid';
    END IF;
    v_snapshot:=platform.analytics_v1_adoption_public_bridge_v1(
      'project',p_tenant_id,p_legacy_endpoint_id,p_target);
    IF (v_snapshot->>'publication_exists')::boolean IS NOT TRUE OR
       (v_snapshot->>'route_exists')::boolean IS NOT TRUE OR
       v_snapshot->>'publication_public_id' IS DISTINCT FROM
         v_pointer->>'public_id' OR
       (v_snapshot->>'publication_release_id')::uuid IS DISTINCT FROM
         (v_pointer->>'release_id')::uuid OR
       (v_snapshot->>'publication_resource_version_id')::uuid IS DISTINCT FROM
         (v_pointer->>'resource_version_id')::uuid OR
       v_snapshot->>'publication_audience'<>'public' OR
       v_snapshot->>'publication_state'<>'active' OR
       v_snapshot->>'route_public_id' IS DISTINCT FROM v_route->>'public_id' OR
       (v_snapshot->>'route_release_id')::uuid IS DISTINCT FROM
         (v_route->>'release_id')::uuid OR
       v_snapshot->>'route_state'<>'active' OR
       (v_snapshot->>'hit_count')::bigint<>0 OR
       v_snapshot->'last_hit_at'<>'null'::jsonb OR
       (v_snapshot->>'consumers_acked_at')::timestamptz IS DISTINCT FROM
         (v_route->>'consumers_acked_at')::timestamptz OR
       (v_snapshot->>'remove_after')::timestamptz IS DISTINCT FROM
         (v_route->>'remove_after')::timestamptz OR
       v_snapshot->>'consumer_inventory_sha256' IS DISTINCT FROM
         v_route->>'consumer_inventory_sha256' OR
       v_snapshot->>'legacy_tenant_identity' IS DISTINCT FROM p_tenant_id::text OR
       v_snapshot->>'method' IS DISTINCT FROM v_route->>'method' OR
       v_snapshot->>'path' IS DISTINCT FROM v_route->>'path' OR
       v_snapshot->>'slug' IS DISTINCT FROM v_route->>'slug' THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption public projection differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_project_public_v1(
    uuid,uuid,jsonb,jsonb
) FROM PUBLIC;
