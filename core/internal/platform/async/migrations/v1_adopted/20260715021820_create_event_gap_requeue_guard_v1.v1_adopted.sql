-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_is_gap_requeue_v1(
    p_old platform.outbox,
    p_new platform.outbox,
    p_now timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN (
SELECT
    p_old.delivery_state = 'delivered'
    AND p_new.delivery_state = 'pending'
    AND p_old.payload_canonical IS NOT NULL
    AND p_old.payload_redacted_at IS NULL
    AND p_new.payload_canonical IS NOT DISTINCT FROM p_old.payload_canonical
    AND p_new.payload_redacted_at IS NULL
    AND p_new.delivery_attempts = p_old.delivery_attempts
    AND p_new.lease_generation = p_old.lease_generation
    AND p_new.lease_owner_id IS NULL
    AND p_new.lease_heartbeat_at IS NULL
    AND p_new.lease_until IS NULL
    AND p_new.next_attempt_at <= p_now
    AND p_new.next_attempt_at >= p_old.updated_at
    AND p_new.confirmed_at IS NULL
    AND p_new.delivered_at IS NULL
    AND p_new.confirm_sha256 IS NULL
    AND p_new.last_error_code IS NULL
    AND p_new.last_error_sha256 IS NULL
    AND p_new.parked_at IS NOT DISTINCT FROM p_old.parked_at
    AND p_new.parked_code IS NOT DISTINCT FROM p_old.parked_code
    AND p_new.parked_current_group_generation
        IS NOT DISTINCT FROM p_old.parked_current_group_generation
    AND p_new.parked_current_cutover_generation
        IS NOT DISTINCT FROM p_old.parked_current_cutover_generation
    AND EXISTS (
        SELECT 1
        FROM platform.jobs AS job
        JOIN platform.dead_letters AS root
          ON job.payload_ref = 'deadletter:event-gap:' || root.id::text
        JOIN platform.dead_letters AS action
          ON action.id = job.id AND action.root_id = root.id
        WHERE job.state = 'running'
          AND job.kind = 'platform_async.event_gap_reconcile.' ||
              job.authority_domain
          AND job.effect_class = 'database_transactional'
          AND root.record_kind = 'root' AND root.source_kind = 'event'
          AND root.source_id = p_old.event_id
          AND root.reason_class = 'sequence_gap'
          AND action.record_kind = 'retry_action'
          AND action.source_id = p_old.event_id
          AND action.recorded_at = p_new.next_attempt_at
          AND platform.platform_async_event_gap_job_ready_v1(job.id)
    )
);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_is_gap_requeue_v1(
    platform.outbox,platform.outbox,timestamptz
) FROM PUBLIC;
