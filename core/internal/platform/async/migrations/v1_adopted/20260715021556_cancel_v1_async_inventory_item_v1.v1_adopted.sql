-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_cancel_v1_inventory_item_v1(
    p_run_id uuid,
    p_inventory_class text,
    p_source_name text,
    p_source_item_id text,
    p_item_sha256 text,
    p_reason_code text,
    p_reason text,
    p_ticket text,
    p_disposed_at timestamptz,
    p_receipt_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run_state text;
    v_item platform.async_v1_adoption_items%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5650',
            MESSAGE = 'V1 async item cancellation requires adopter authority';
    END IF;
    SELECT state INTO v_run_state FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id FOR UPDATE;
    SELECT * INTO v_item FROM platform.async_v1_adoption_items
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
      AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
      AND source_name OPERATOR(pg_catalog.=) p_source_name
      AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
    FOR SHARE;
    IF v_run_state IS DISTINCT FROM 'disposing' OR NOT FOUND OR
       v_item.item_sha256 IS DISTINCT FROM p_item_sha256 OR
       p_reason_code IS NULL OR p_reason IS NULL OR p_ticket IS NULL OR
       p_disposed_at < v_item.captured_at OR
       p_disposed_at > pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P5651',
            MESSAGE = 'V1 async cancellation evidence is invalid';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.async_v1_adoption_dispositions
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
          AND source_name OPERATOR(pg_catalog.=) p_source_name
          AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
          AND disposition OPERATOR(pg_catalog.=) 'cancelled'
          AND reason_code OPERATOR(pg_catalog.=) p_reason_code
          AND reason OPERATOR(pg_catalog.=) p_reason
          AND ticket OPERATOR(pg_catalog.=) p_ticket
          AND receipt_sha256 OPERATOR(pg_catalog.=) p_receipt_sha256
    ) THEN
        RETURN;
    END IF;
    INSERT INTO platform.async_v1_adoption_dispositions (
        run_id, inventory_class, source_name, source_item_id,
        item_sha256, idempotency_identity, disposition,
        reason_code, reason, ticket, disposed_at, receipt_sha256
    ) VALUES (
        p_run_id, p_inventory_class, p_source_name, p_source_item_id,
        p_item_sha256, v_item.idempotency_identity, 'cancelled',
        p_reason_code, p_reason, p_ticket, p_disposed_at, p_receipt_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_cancel_v1_inventory_item_v1(
    uuid,text,text,text,text,text,text,text,timestamptz,text
) FROM PUBLIC;
