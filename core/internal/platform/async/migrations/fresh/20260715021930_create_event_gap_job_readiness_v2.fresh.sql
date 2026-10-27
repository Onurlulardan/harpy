-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_event_gap_job_ready_v2(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_event platform.outbox%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5690',
            MESSAGE = 'event gap readiness authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job WHERE job.id = p_job_id;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    IF pg_catalog.left(v_job.kind,
            pg_catalog.length('platform_async.event_gap_reconcile.')) <>
            'platform_async.event_gap_reconcile.' THEN
        RETURN true;
    END IF;
    IF v_job.kind IS DISTINCT FROM
        ('platform_async.event_gap_reconcile.' || v_job.authority_domain) THEN
        RETURN false;
    END IF;
    IF platform.platform_async_event_gap_job_ready_v1(p_job_id) THEN
        RETURN true;
    END IF;
    SELECT event.* INTO v_event
    FROM platform.dead_letters AS root
    JOIN platform.outbox AS event ON event.event_id = root.source_id
    CROSS JOIN LATERAL (
        SELECT pg_catalog.convert_from(root.payload_canonical,'UTF8')::jsonb AS value
    ) AS gap
    WHERE v_job.payload_ref = 'deadletter:event-gap:' || root.id::text
      AND root.record_kind = 'root' AND root.source_kind = 'event'
      AND root.reason_class = 'sequence_gap'
      AND root.reason_code = 'EVENT_SEQUENCE_GAP'
      AND root.payload_canonical IS NOT NULL
      AND root.payload_redacted_at IS NULL
      AND ROW(root.tenant_id,root.scope_kind,root.scope_id,
          root.authority_group_id,root.authority_group_generation,
          root.authority_domain,root.cutover_generation,root.writer_origin,
          root.source_sha256)
          = ROW(v_job.tenant_id,v_job.scope_kind,v_job.scope_id,
              v_job.authority_group_id,v_job.authority_group_generation,
              v_job.authority_domain,v_job.cutover_generation,'v2',
              v_job.payload_sha256)
      AND gap.value->>'contract_version' = 'event-sequence-gap.v1'
      AND gap.value->>'consumer' = root.consumer_name
      AND (gap.value->>'consumer_version')::integer = root.consumer_version
      AND gap.value->'envelope'->>'event_id' = event.event_id::text
      AND (gap.value->'envelope'->>'event_sequence')::bigint = event.event_sequence
      AND gap.value->'envelope'->>'event_type' = event.event_type
      AND (gap.value->'envelope'->>'event_version')::integer = event.event_version
      AND gap.value->'envelope'->>'aggregate_type' = event.aggregate_type
      AND gap.value->'envelope'->>'aggregate_id' = event.aggregate_id::text
      AND gap.value->'envelope'->>'payload_sha256' = event.payload_sha256
      AND event.payload_canonical IS NOT NULL
      AND event.payload_redacted_at IS NULL
      AND gap.value->'envelope'->'payload' =
          pg_catalog.convert_from(event.payload_canonical,'UTF8')::jsonb;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    RETURN NOT platform.platform_async_outbox_authority_current_v1(v_event);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_event_gap_job_ready_v2(uuid)
FROM PUBLIC;
