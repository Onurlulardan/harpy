-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_adoption_accept_v1(
    p_scope_schema name, p_authority_id uuid, p_tenant_id uuid,
    p_run_id uuid, p_cutover_generation bigint, p_source_count bigint,
    p_adopted_count bigint, p_retired_count bigint,
    p_root_sha256 bytea, p_accepted_at timestamptz
)
RETURNS TABLE (
    run_id uuid, tenant_id uuid, cutover_generation bigint,
    source_count bigint, adopted_count bigint, retired_count bigint,
    root_sha256 bytea, accepted_at timestamptz, replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_inserted boolean;
    v_source bigint;
    v_adopted bigint;
    v_retired bigint;
    v_blocked bigint;
    v_root bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS NULL OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema()::text THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics adoption acceptance authority is invalid';
    END IF;
    PERFORM platform.analytics_assert_tenant_authority_v1(
        p_authority_id, p_tenant_id
    );
    SELECT pg_catalog.count(*),
           pg_catalog.count(*) FILTER (WHERE item.disposition = 'adopted'),
           pg_catalog.count(*) FILTER (WHERE item.disposition = 'retired'),
           pg_catalog.count(*) FILTER (WHERE item.disposition = 'blocked')
      INTO v_source, v_adopted, v_retired, v_blocked
    FROM analytics_v1_adoption_items AS item
    WHERE item.run_id = p_run_id;
    SELECT pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(
            item.legacy_endpoint_id::text || '|' ||
            pg_catalog.encode(item.source_sha256, 'hex') || '|' ||
            item.disposition || '|' || COALESCE(item.reason_code, '') || '|' ||
            COALESCE(item.query_id::text, '') || '|' ||
            COALESCE(item.query_version_id::text, '') || '|' ||
            COALESCE(item.publication_release_id::text, '') || '|' ||
            COALESCE(pg_catalog.encode(item.target_sha256, 'hex'), ''),
            E'\n' ORDER BY item.legacy_endpoint_id
        ), ''), 'UTF8'
    )) INTO v_root
    FROM analytics_v1_adoption_items AS item
    WHERE item.run_id = p_run_id;
    IF v_blocked <> 0 OR v_source IS DISTINCT FROM p_source_count OR
       v_adopted IS DISTINCT FROM p_adopted_count OR
       v_retired IS DISTINCT FROM p_retired_count OR
       p_source_count IS DISTINCT FROM p_adopted_count + p_retired_count OR
       v_root IS DISTINCT FROM p_root_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104',
            MESSAGE = 'analytics adoption acceptance is blocked';
    END IF;
    INSERT INTO analytics_v1_adoption_acceptance_receipts (
        run_id, tenant_id, cutover_generation, source_count,
        adopted_count, retired_count, root_sha256, accepted_at
    ) VALUES (
        p_run_id, p_tenant_id, p_cutover_generation, p_source_count,
        p_adopted_count, p_retired_count, p_root_sha256, p_accepted_at
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM analytics_v1_adoption_acceptance_receipts AS receipt
        WHERE receipt.run_id = p_run_id AND receipt.tenant_id = p_tenant_id
          AND receipt.cutover_generation = p_cutover_generation
          AND receipt.source_count = p_source_count
          AND receipt.adopted_count = p_adopted_count
          AND receipt.retired_count = p_retired_count
          AND receipt.root_sha256 IS NOT DISTINCT FROM p_root_sha256
          AND receipt.accepted_at IS NOT DISTINCT FROM p_accepted_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6103',
            MESSAGE = 'analytics adoption acceptance replay differs';
    END IF;
    RETURN QUERY
    SELECT receipt.run_id, receipt.tenant_id, receipt.cutover_generation,
           receipt.source_count, receipt.adopted_count, receipt.retired_count,
           receipt.root_sha256, receipt.accepted_at,
           NOT COALESCE(v_inserted, false)
    FROM analytics_v1_adoption_acceptance_receipts AS receipt
    WHERE receipt.run_id = p_run_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_adoption_accept_v1(
    name,uuid,uuid,uuid,bigint,bigint,bigint,bigint,bytea,timestamptz
) FROM PUBLIC;
