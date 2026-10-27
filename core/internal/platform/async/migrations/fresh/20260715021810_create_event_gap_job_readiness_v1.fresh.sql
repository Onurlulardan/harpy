-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_event_gap_job_ready_v1(p_job_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN (
SELECT CASE
    WHEN pg_catalog.left(
        job.kind, pg_catalog.length('platform_async.event_gap_reconcile.')
    ) <> 'platform_async.event_gap_reconcile.' THEN true
    WHEN job.kind IS DISTINCT FROM
        ('platform_async.event_gap_reconcile.' || job.authority_domain) THEN false
    ELSE EXISTS (
        SELECT 1
        FROM platform.dead_letters AS root
        JOIN platform.outbox AS event ON event.event_id = root.source_id
        CROSS JOIN LATERAL (
            SELECT pg_catalog.convert_from(root.payload_canonical, 'UTF8')::jsonb AS value
        ) AS gap
        WHERE root.id::text = pg_catalog.substr(
                job.payload_ref, pg_catalog.length('deadletter:event-gap:') + 1
            )
          AND job.payload_ref = 'deadletter:event-gap:' || root.id::text
          AND root.record_kind = 'root' AND root.source_kind = 'event'
          AND root.reason_class = 'sequence_gap'
          AND root.reason_code = 'EVENT_SEQUENCE_GAP'
          AND root.payload_canonical IS NOT NULL
          AND root.payload_redacted_at IS NULL
          AND ROW(root.tenant_id,root.scope_kind,root.scope_id,
              root.authority_group_id,root.authority_group_generation,
              root.authority_domain,root.cutover_generation,root.writer_origin,
              root.source_sha256)
              = ROW(job.tenant_id,job.scope_kind,job.scope_id,
                  job.authority_group_id,job.authority_group_generation,
                  job.authority_domain,job.cutover_generation,'v2',job.payload_sha256)
          AND gap.value->>'contract_version' = 'event-sequence-gap.v1'
          AND gap.value->>'consumer' = root.consumer_name
          AND (gap.value->>'consumer_version')::integer = root.consumer_version
          AND (gap.value->>'received_sequence')::bigint = event.event_sequence
          AND gap.value->'envelope'->>'event_id' = event.event_id::text
          AND gap.value->'envelope'->>'event_type' = event.event_type
          AND (gap.value->'envelope'->>'event_version')::integer = event.event_version
          AND gap.value->'envelope'->>'aggregate_type' = event.aggregate_type
          AND gap.value->'envelope'->>'aggregate_id' = event.aggregate_id::text
          AND (gap.value->'envelope'->>'event_sequence')::bigint = event.event_sequence
          AND gap.value->'envelope'->>'payload_sha256' = event.payload_sha256
          AND event.payload_canonical IS NOT NULL
          AND gap.value->'envelope'->'payload' =
              pg_catalog.convert_from(event.payload_canonical, 'UTF8')::jsonb
          AND (
              EXISTS (
                  SELECT 1
                  FROM platform.inbox AS receipt
                  WHERE receipt.record_kind = 'event_receipt'
                    AND receipt.consumer_name = root.consumer_name
                    AND receipt.consumer_version = root.consumer_version
                    AND receipt.scope_kind = root.scope_kind
                    AND receipt.scope_id = root.scope_id
                    AND receipt.event_id = root.source_id
              ) OR (
                  SELECT COALESCE(pg_catalog.max(mark.through_sequence), 0)
                  FROM platform.inbox AS mark
                  WHERE mark.consumer_name = root.consumer_name
                    AND mark.consumer_version = root.consumer_version
                    AND mark.scope_kind = root.scope_kind
                    AND mark.scope_id = root.scope_id
                    AND mark.aggregate_type = event.aggregate_type
                    AND mark.aggregate_id = event.aggregate_id
              ) = event.event_sequence - 1
          )
    )
END
FROM platform.jobs AS job
WHERE job.id = p_job_id
);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_event_gap_job_ready_v1(uuid)
FROM PUBLIC;
