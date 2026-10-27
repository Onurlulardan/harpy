-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_admit_v1(
    p_tenant_id uuid, p_release_id uuid, p_generation bigint,
    p_client_ip inet, p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_window timestamptz; v_release_count integer; v_tenant_count bigint;
    v_lease_id uuid;
BEGIN
    IF p_tenant_id IS NULL OR pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_generation < 1 OR p_client_ip IS NULL OR p_occurred_at IS NULL OR
       NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics public admission is invalid';
    END IF;
    PERFORM 1 FROM platform.analytics_publications AS publication
    WHERE publication.tenant_id = p_tenant_id AND publication.release_id = p_release_id
      AND publication.generation = p_generation AND publication.state = 'active' FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics public route was not found';
    END IF;
    v_window := pg_catalog.date_trunc('minute', p_occurred_at);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:public-limit:' || p_tenant_id::text || ':' || v_window::text, 6110));
    DELETE FROM platform.analytics_public_concurrency_leases WHERE expires_at <= p_occurred_at;
    SELECT pg_catalog.count(*)::integer INTO v_release_count
    FROM platform.analytics_public_concurrency_leases
    WHERE release_id = p_release_id AND generation = p_generation;
    IF v_release_count >= 10 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6105', MESSAGE = 'analytics public concurrency limit reached';
    END IF;
    SELECT COALESCE(pg_catalog.sum(request_count), 0) INTO v_tenant_count
    FROM platform.analytics_public_rate_windows
    WHERE tenant_id = p_tenant_id AND window_started_at = v_window;
    IF v_tenant_count >= 600 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6105', MESSAGE = 'analytics tenant public rate limit reached';
    END IF;
    INSERT INTO platform.analytics_public_rate_windows (
        tenant_id, release_id, generation, client_ip, window_started_at, request_count
    ) VALUES (p_tenant_id, p_release_id, p_generation, p_client_ip, v_window, 1)
    ON CONFLICT (tenant_id, release_id, generation, client_ip, window_started_at)
    DO UPDATE SET request_count = platform.analytics_public_rate_windows.request_count + 1
    WHERE platform.analytics_public_rate_windows.request_count < 60;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6105', MESSAGE = 'analytics public client rate limit reached';
    END IF;
    v_lease_id := pg_catalog.uuidv7();
    INSERT INTO platform.analytics_public_concurrency_leases VALUES (
        v_lease_id, p_tenant_id, p_release_id, p_generation, p_occurred_at + interval '30 seconds');
    RETURN v_lease_id;
END;
$function$;

CREATE FUNCTION platform.analytics_public_release_v1(
    p_lease_id uuid, p_tenant_id uuid, p_release_id uuid, p_generation bigint
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF pg_catalog.uuid_extract_version(p_lease_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics public lease is invalid';
    END IF;
    DELETE FROM platform.analytics_public_concurrency_leases
    WHERE lease_id = p_lease_id AND tenant_id = p_tenant_id
      AND release_id = p_release_id AND generation = p_generation;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_admit_v1(
    uuid,uuid,bigint,inet,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.analytics_public_release_v1(
    uuid,uuid,uuid,bigint
) FROM PUBLIC;
