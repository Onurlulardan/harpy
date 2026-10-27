-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_seal_v1_inventory_v1(
    p_run_id uuid,
    p_inventory_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.async_v1_adoption_runs%ROWTYPE;
    v_item_count bigint;
    v_inventory_sha256 text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5641',
            MESSAGE = 'V1 async inventory seal requires adopter authority';
    END IF;
    SELECT * INTO v_run
    FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
    FOR UPDATE;
    IF NOT FOUND OR v_run.state NOT IN ('inventory', 'disposing') OR
       v_run.inventory_sha256 IS DISTINCT FROM p_inventory_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5642',
            MESSAGE = 'V1 async inventory seal identity differs';
    END IF;
    SELECT pg_catalog.count(*),
           pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
               COALESCE(pg_catalog.string_agg(item_sha256, E'\n' ORDER BY item_sha256), ''),
               'UTF8'
           )), 'hex')
    INTO v_item_count, v_inventory_sha256
    FROM platform.async_v1_adoption_items
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    IF v_item_count IS DISTINCT FROM v_run.item_count OR
       v_inventory_sha256 IS DISTINCT FROM v_run.inventory_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5643',
            MESSAGE = 'V1 async inventory count or checksum differs';
    END IF;
    IF v_run.state OPERATOR(pg_catalog.=) 'inventory' THEN
        UPDATE platform.async_v1_adoption_runs
        SET state = 'disposing'
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_seal_v1_inventory_v1(uuid,text)
FROM PUBLIC;
