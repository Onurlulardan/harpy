-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_adopt_v1_inventory_item_v1(
    p_run_id uuid,
    p_inventory_class text,
    p_source_name text,
    p_source_item_id text,
    p_item_sha256 text,
    p_job_id uuid,
    p_disposed_at timestamptz,
    p_receipt_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.async_v1_adoption_runs%ROWTYPE;
    v_item platform.async_v1_adoption_items%ROWTYPE;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5646',
            MESSAGE = 'V1 async item adoption requires adopter authority';
    END IF;
    SELECT * INTO v_run FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id FOR UPDATE;
    SELECT * INTO v_item FROM platform.async_v1_adoption_items
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
      AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
      AND source_name OPERATOR(pg_catalog.=) p_source_name
      AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
    FOR SHARE;
    IF v_run.state IS DISTINCT FROM 'disposing' OR NOT FOUND OR
       v_item.item_sha256 IS DISTINCT FROM p_item_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5647',
            MESSAGE = 'V1 async adoption item is not sealed';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.async_v1_adoption_dispositions
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
          AND source_name OPERATOR(pg_catalog.=) p_source_name
          AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
          AND disposition OPERATOR(pg_catalog.=) 'adopted'
          AND target_job_id OPERATOR(pg_catalog.=) p_job_id
          AND receipt_sha256 OPERATOR(pg_catalog.=) p_receipt_sha256
    ) THEN
        RETURN;
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.async_v1_adoption_dispositions
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
          AND source_name OPERATOR(pg_catalog.=) p_source_name
          AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5648',
            MESSAGE = 'V1 async adoption disposition already differs';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    IF p_disposed_at < v_item.captured_at OR p_disposed_at > v_now THEN
        RAISE EXCEPTION USING ERRCODE = 'P5649',
            MESSAGE = 'V1 async adoption disposition time is invalid';
    END IF;
    INSERT INTO platform.async_v1_adoption_job_bindings (
        run_id, inventory_class, source_name, source_item_id, item_sha256, job_id
    ) VALUES (
        p_run_id, p_inventory_class, p_source_name, p_source_item_id,
        p_item_sha256, p_job_id
    );
    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        kind, payload_revision, payload_ref, payload_sha256,
        dedupe_key, effect_class, effect_idempotency_key,
        priority, run_at, state, attempts, max_attempts,
        lease_generation, created_at, updated_at
    ) VALUES (
        p_job_id,
        CASE WHEN v_run.scope_kind = 'tenant' THEN v_run.scope_id ELSE NULL END,
        v_run.scope_kind, v_run.scope_id, v_run.group_id,
        v_run.target_group_generation, v_run.domain, v_run.target_generation, 'v1',
        v_item.target_kind, v_item.target_payload_revision,
        v_item.target_payload_ref, v_item.target_payload_sha256,
        v_item.idempotency_identity, v_item.target_effect_class,
        v_item.target_effect_idempotency_key, v_item.target_priority,
        v_item.target_run_at, 'queued', 0, v_item.target_max_attempts,
        0, v_now, v_now
    );
    INSERT INTO platform.async_v1_adoption_dispositions (
        run_id, inventory_class, source_name, source_item_id,
        item_sha256, idempotency_identity, disposition,
        target_job_id, disposed_at, receipt_sha256
    ) VALUES (
        p_run_id, p_inventory_class, p_source_name, p_source_item_id,
        p_item_sha256, v_item.idempotency_identity, 'adopted',
        p_job_id, p_disposed_at, p_receipt_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_adopt_v1_inventory_item_v1(
    uuid,text,text,text,text,uuid,timestamptz,text
) FROM PUBLIC;
