-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_v1_zero_inventory_v1(
    p_run_id uuid,
    p_inventory_sha256 text,
    p_queue_count bigint,
    p_timer_count bigint,
    p_future_count bigint,
    p_running_count bigint,
    p_observed_at timestamptz,
    p_receipt_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.async_v1_adoption_runs%ROWTYPE;
    v_enqueue_count bigint;
    v_scheduler_count bigint;
    v_consumer_count bigint;
    v_component_sha256 text;
    v_last_restart timestamptz;
    v_disposition_count bigint;
    v_inserted boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_cluster_admin' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5655',
            MESSAGE = 'V1 zero inventory requires cluster authority';
    END IF;
    SELECT * INTO v_run FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id FOR UPDATE;
    IF NOT FOUND OR v_run.state IS DISTINCT FROM 'disposing' OR
       v_run.inventory_sha256 IS DISTINCT FROM p_inventory_sha256 OR
       p_queue_count IS DISTINCT FROM 0 OR p_timer_count IS DISTINCT FROM 0 OR
       p_future_count IS DISTINCT FROM 0 OR p_running_count IS DISTINCT FROM 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5656',
            MESSAGE = 'V1 async inventory is not zero';
    END IF;
    SELECT pg_catalog.count(*) FILTER (WHERE fence_kind = 'enqueue'),
           pg_catalog.count(*) FILTER (WHERE fence_kind = 'scheduler'),
           pg_catalog.count(*) FILTER (WHERE fence_kind = 'consumer'),
           pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
               COALESCE(pg_catalog.string_agg(
                   receipt_sha256, E'\n' ORDER BY receipt_sha256
               ), ''), 'UTF8'
           )), 'hex'),
           pg_catalog.max(restart_verified_at)
    INTO v_enqueue_count, v_scheduler_count, v_consumer_count,
         v_component_sha256, v_last_restart
    FROM platform.async_v1_restart_fence_receipts
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    IF v_enqueue_count IS DISTINCT FROM v_run.expected_enqueue_fences OR
       v_scheduler_count IS DISTINCT FROM v_run.expected_scheduler_fences OR
       v_consumer_count IS DISTINCT FROM v_run.expected_consumer_fences OR
       v_component_sha256 IS DISTINCT FROM v_run.component_manifest_sha256 OR
       v_last_restart IS NULL OR p_observed_at < v_last_restart OR
       p_observed_at < v_run.captured_at OR
       p_observed_at > pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P5657',
            MESSAGE = 'V1 enqueue scheduler or consumer restart fence is incomplete';
    END IF;
    SELECT pg_catalog.count(*) INTO v_disposition_count
    FROM platform.async_v1_adoption_dispositions
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id;
    IF v_disposition_count IS DISTINCT FROM v_run.item_count THEN
        RAISE EXCEPTION USING ERRCODE = 'P5658',
            MESSAGE = 'V1 async inventory has unresolved items';
    END IF;
    INSERT INTO platform.async_v1_zero_inventory_receipts (
        run_id, inventory_contract_version, inventory_sha256,
        queue_count, timer_count, future_count, running_count,
        observed_at, receipt_sha256
    ) VALUES (
        p_run_id, 'async-adoption-zero-inventory.v1', p_inventory_sha256,
        p_queue_count, p_timer_count, p_future_count, p_running_count,
        p_observed_at, p_receipt_sha256
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM platform.async_v1_zero_inventory_receipts
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND receipt_sha256 OPERATOR(pg_catalog.=) p_receipt_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5659',
            MESSAGE = 'V1 zero inventory replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_record_v1_zero_inventory_v1(
    uuid,text,bigint,bigint,bigint,bigint,timestamptz,text
) FROM PUBLIC;
