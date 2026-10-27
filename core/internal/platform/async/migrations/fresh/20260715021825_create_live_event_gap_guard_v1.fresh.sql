-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_has_live_event_gap_root_v1(
    p_root platform.dead_letters
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN p_root.record_kind = 'root'
    AND p_root.source_kind = 'event'
    AND p_root.reason_class = 'sequence_gap'
    AND EXISTS (
        SELECT 1
        FROM platform.jobs AS job
        WHERE job.payload_ref = 'deadletter:event-gap:' || p_root.id::text
          AND job.payload_sha256 = p_root.payload_sha256
          AND job.kind = 'platform_async.event_gap_reconcile.' ||
              job.authority_domain
          AND job.state IN ('queued','running')
    );
END;
$function$;

CREATE FUNCTION platform.platform_async_has_live_event_gap_v1(
    p_event platform.outbox
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN EXISTS (
    SELECT 1
    FROM platform.dead_letters AS root
    WHERE root.source_id = p_event.event_id
      AND root.tenant_id = p_event.tenant_id
      AND root.scope_kind = p_event.scope_kind
      AND root.scope_id = p_event.scope_id
      AND platform.platform_async_has_live_event_gap_root_v1(root)
);
END;
$function$;

CREATE FUNCTION platform.platform_async_has_live_stale_job_root_v1(
    p_root platform.dead_letters
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN p_root.record_kind = 'root'
    AND p_root.source_kind = 'job'
    AND p_root.reason_class = 'stale_generation'
    AND EXISTS (
        SELECT 1
        FROM platform.jobs AS job
        WHERE job.id = p_root.source_id
          AND job.state IN ('queued','running')
          AND job.parked_at IS NOT NULL
          AND job.parked_code = 'JOB_AUTHORITY_STALE'
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_has_live_event_gap_root_v1(platform.dead_letters)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_has_live_event_gap_v1(platform.outbox)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_has_live_stale_job_root_v1(platform.dead_letters)
FROM PUBLIC;
