-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_consumer_sha256_v1(p_tenant_id uuid)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS tenant
        WHERE tenant.singleton AND tenant.tenant_id=p_tenant_id
          AND tenant.schema_name=pg_catalog.current_schema()
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption consumer evidence denied';
    END IF;
    IF EXISTS (
      SELECT 1 FROM catalog_legacy_route_hits AS route
      LEFT JOIN LATERAL (
        SELECT pg_catalog.count(*) AS count
        FROM catalog_legacy_consumer_acks AS acknowledgement
        WHERE acknowledgement.route_key=route.route_key
      ) AS actual ON true
      WHERE route.retired_at IS NULL OR route.retired_at<route.remove_after OR
        route.consumer_ack_count<>route.required_consumer_count OR
        route.consumer_ack_count<>actual.count OR
        (route.last_hit_at IS NOT NULL AND route.last_hit_at>=route.remove_after) OR
        (route.hit_count=0) IS DISTINCT FROM (route.last_hit_at IS NULL)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption consumers are not retired';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.int4send(pg_catalog.octet_length(row.kind))||
      pg_catalog.convert_to(row.kind,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(row.identity))||
      pg_catalog.convert_to(row.identity,'UTF8')||
      pg_catalog.int4send(pg_catalog.octet_length(row.document))||
      pg_catalog.convert_to(row.document,'UTF8'),
      ''::bytea ORDER BY row.kind,row.identity),''::bytea)),'hex')
    INTO v_digest FROM (
      SELECT 'route' AS kind,route.route_key AS identity,to_jsonb(route)::text AS document
      FROM catalog_legacy_route_hits AS route
      UNION ALL
      SELECT 'ack',ack.route_key||':'||ack.consumer_id,to_jsonb(ack)::text
      FROM catalog_legacy_consumer_acks AS ack
    ) AS row;
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_consumer_sha256_v1(uuid) FROM PUBLIC;
