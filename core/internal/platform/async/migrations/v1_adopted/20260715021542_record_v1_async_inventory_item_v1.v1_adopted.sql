-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_v1_inventory_item_v1(
    p_run_id uuid,
    p_inventory_class text,
    p_source_surface text,
    p_source_name text,
    p_source_item_id text,
    p_source_contract_version text,
    p_source_revision integer,
    p_source_payload_sha256 text,
    p_transform_version text,
    p_idempotency_identity text,
    p_target_kind text,
    p_target_payload_revision integer,
    p_target_payload_ref text,
    p_target_payload_sha256 text,
    p_target_run_at timestamptz,
    p_target_priority integer,
    p_target_max_attempts integer,
    p_target_effect_class text,
    p_target_effect_idempotency_key text,
    p_captured_at timestamptz,
    p_item_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_manifest_captured_at timestamptz;
    v_inserted boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5637',
            MESSAGE = 'V1 async inventory requires adopter authority';
    END IF;
    SELECT captured_at INTO v_manifest_captured_at
    FROM platform.async_v1_adoption_runs
    WHERE run_id OPERATOR(pg_catalog.=) p_run_id
      AND state OPERATOR(pg_catalog.=) 'inventory'
    FOR UPDATE;
    IF NOT FOUND OR p_captured_at > v_manifest_captured_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5638',
            MESSAGE = 'V1 async inventory run is not open';
    END IF;
    IF (p_source_name OPERATOR(pg_catalog.=) 'workflow.cron.scheduler' AND NOT (
            p_inventory_class OPERATOR(pg_catalog.=) 'timer' AND
            p_source_surface OPERATOR(pg_catalog.=) 'scheduler' AND
            p_source_contract_version OPERATOR(pg_catalog.=)
                'v1.workflow-minute-scheduler.v1'
       )) OR
       (p_source_name OPERATOR(pg_catalog.=) 'schedule.events' AND NOT (
            p_inventory_class OPERATOR(pg_catalog.=) 'queue' AND
            p_source_surface OPERATOR(pg_catalog.=) 'queue' AND
            p_source_contract_version OPERATOR(pg_catalog.=) 'v1.queue-message.v1'
       )) OR
       (p_source_name OPERATOR(pg_catalog.=) 'workflow.delay' AND NOT (
            p_inventory_class OPERATOR(pg_catalog.=) 'timer' AND
            p_source_surface OPERATOR(pg_catalog.=) 'delayed_plugin' AND
            p_source_contract_version OPERATOR(pg_catalog.=)
                'v1.workflow-delayed-plugin.v1'
       )) OR
       (p_source_name OPERATOR(pg_catalog.=) 'workflow.webhook.retry' AND NOT (
            p_inventory_class OPERATOR(pg_catalog.=) 'timer' AND
            p_source_surface OPERATOR(pg_catalog.=) 'ttl_queue' AND
            p_source_contract_version OPERATOR(pg_catalog.=)
                'v1.workflow-webhook-ttl-not-delay.v1'
       )) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5639',
            MESSAGE = 'known V1 async source semantics differ';
    END IF;
    INSERT INTO platform.async_v1_adoption_items (
        run_id, inventory_class, source_surface, source_name, source_item_id,
        source_contract_version, source_revision, source_payload_sha256,
        transform_version, idempotency_identity, target_kind,
        target_payload_revision, target_payload_ref, target_payload_sha256,
        target_run_at, target_priority, target_max_attempts,
        target_effect_class, target_effect_idempotency_key,
        captured_at, item_sha256
    ) VALUES (
        p_run_id, p_inventory_class, p_source_surface, p_source_name, p_source_item_id,
        p_source_contract_version, p_source_revision, p_source_payload_sha256,
        p_transform_version, p_idempotency_identity, p_target_kind,
        p_target_payload_revision, p_target_payload_ref, p_target_payload_sha256,
        p_target_run_at, p_target_priority, p_target_max_attempts,
        p_target_effect_class, p_target_effect_idempotency_key,
        p_captured_at, p_item_sha256
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM platform.async_v1_adoption_items
        WHERE run_id OPERATOR(pg_catalog.=) p_run_id
          AND inventory_class OPERATOR(pg_catalog.=) p_inventory_class
          AND source_name OPERATOR(pg_catalog.=) p_source_name
          AND source_item_id OPERATOR(pg_catalog.=) p_source_item_id
          AND item_sha256 OPERATOR(pg_catalog.=) p_item_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5640',
            MESSAGE = 'V1 async inventory item replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_record_v1_inventory_item_v1(
    uuid,text,text,text,text,text,integer,text,text,text,text,integer,text,text,
    timestamptz,integer,integer,text,text,timestamptz,text
) FROM PUBLIC;
