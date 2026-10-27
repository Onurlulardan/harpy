-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_v1_adoption_public_bridge_v1(
    p_action text,p_tenant_id uuid,p_legacy_endpoint_id uuid,p_target jsonb
) RETURNS jsonb
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_pointer jsonb;
    v_route jsonb;
    v_snapshot jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_action NOT IN ('count','inspect','project') OR p_tenant_id IS NULL OR
       (p_action<>'count' AND p_legacy_endpoint_id IS NULL) OR
       (p_action='project' AND p_target IS NULL) THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption public bridge denied';
    END IF;
    IF p_action='count' THEN
      SELECT pg_catalog.jsonb_build_object(
        'publication_count',(SELECT pg_catalog.count(*)
          FROM platform.analytics_publications WHERE tenant_id=p_tenant_id),
        'route_count',(SELECT pg_catalog.count(*)
          FROM platform.analytics_legacy_public_routes WHERE tenant_id=p_tenant_id)
      ) INTO v_snapshot;
      RETURN v_snapshot;
    END IF;
    IF p_action='project' THEN
      v_pointer:=p_target->'public_pointer';
      v_route:=p_target->'legacy_route';
      IF (v_pointer->>'tenant_id')::uuid IS DISTINCT FROM p_tenant_id OR
         (v_route->>'tenant_id')::uuid IS DISTINCT FROM p_tenant_id OR
         (v_route->>'legacy_endpoint_id')::uuid
           IS DISTINCT FROM p_legacy_endpoint_id OR
         v_pointer->>'public_id'!~'^[0-9a-f]{32}$' OR
         v_route->>'public_id' IS DISTINCT FROM v_pointer->>'public_id' OR
         (v_route->>'release_id')::uuid IS DISTINCT FROM
           (v_pointer->>'release_id')::uuid OR
         v_pointer->>'state'<>'active' OR v_route->>'state'<>'active' OR
         (v_route->>'hit_count')::bigint<>0 OR
         v_route->'last_hit_at'<>'null'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE='22023',
          MESSAGE='Analytics V1 adoption public bridge input invalid';
      END IF;
      INSERT INTO platform.analytics_publications(
        public_id,tenant_id,release_id,resource_kind,resource_version_id,audience,
        state,generation,capability_sha256,capability_expires_at,activated_at,revoked_at
      ) VALUES (
        pg_catalog.decode(v_pointer->>'public_id','hex'),p_tenant_id,
        (v_pointer->>'release_id')::uuid,v_pointer->>'resource_kind',
        (v_pointer->>'resource_version_id')::uuid,v_pointer->>'audience','active',
        (v_pointer->>'generation')::bigint,NULL,NULL,
        (v_pointer->>'activated_at')::timestamptz,NULL
      ) ON CONFLICT DO NOTHING;
      INSERT INTO platform.analytics_legacy_public_routes(
        legacy_tenant_identity,method,path,slug,tenant_id,legacy_endpoint_id,
        public_id,release_id,state,hit_count,last_hit_at,consumers_acked_at,
        remove_after,consumer_inventory_sha256,retirement_evidence_sha256,
        retirement_observed_at,retired_at
      ) VALUES (
        v_route->>'legacy_tenant_identity',v_route->>'method',v_route->>'path',
        v_route->>'slug',p_tenant_id,p_legacy_endpoint_id,
        pg_catalog.decode(v_route->>'public_id','hex'),
        (v_route->>'release_id')::uuid,'active',0,NULL,
        (v_route->>'consumers_acked_at')::timestamptz,
        (v_route->>'remove_after')::timestamptz,
        pg_catalog.decode(v_route->>'consumer_inventory_sha256','hex'),NULL,NULL,NULL
      ) ON CONFLICT DO NOTHING;
    END IF;
    SELECT pg_catalog.jsonb_build_object(
      'publication_exists',publication.public_id IS NOT NULL,
      'publication_public_id',pg_catalog.encode(publication.public_id,'hex'),
      'publication_release_id',publication.release_id,
      'publication_resource_version_id',publication.resource_version_id,
      'publication_audience',publication.audience,
      'publication_state',publication.state,
      'route_exists',route.legacy_endpoint_id IS NOT NULL,
      'route_public_id',pg_catalog.encode(route.public_id,'hex'),
      'route_release_id',route.release_id,'route_state',route.state,
      'hit_count',route.hit_count,'last_hit_at',route.last_hit_at,
      'consumers_acked_at',route.consumers_acked_at,
      'remove_after',route.remove_after,
      'consumer_inventory_sha256',
        pg_catalog.encode(route.consumer_inventory_sha256,'hex'),
      'legacy_tenant_identity',route.legacy_tenant_identity,
      'method',route.method,'path',route.path,'slug',route.slug
    ) INTO v_snapshot
    FROM (SELECT 1 AS marker) AS seed
    LEFT JOIN platform.analytics_legacy_public_routes AS route
      ON route.tenant_id=p_tenant_id
     AND route.legacy_endpoint_id=p_legacy_endpoint_id
    LEFT JOIN platform.analytics_publications AS publication
      ON publication.public_id=route.public_id
     AND publication.tenant_id=route.tenant_id
     AND publication.release_id=route.release_id;
    RETURN v_snapshot;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_v1_adoption_public_bridge_v1(
    text,uuid,uuid,jsonb
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.analytics_v1_adoption_public_bridge_v1(
    text,uuid,uuid,jsonb
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) FROM forgecrud_adopter;
