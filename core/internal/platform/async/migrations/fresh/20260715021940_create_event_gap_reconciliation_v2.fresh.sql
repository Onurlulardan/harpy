-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_reconcile_event_gap_v2(
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint
)
RETURNS TABLE(
    completed_at timestamptz,requeued boolean,action_id uuid,disposition text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_root platform.dead_letters%ROWTYPE;
    v_event platform.outbox%ROWTYPE;
    v_gap jsonb;
    v_now timestamptz;
    v_error_sha256 text;
    v_count bigint;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5690',
            MESSAGE = 'event gap reconciliation authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id = p_job_id AND job.state = 'running'
      AND job.kind = 'platform_async.event_gap_reconcile.' || job.authority_domain
      AND job.payload_revision = 1
      AND job.effect_class = 'database_transactional'
      AND job.effect_started_at IS NULL AND job.cancel_requested_at IS NULL
      AND job.lease_owner_id = p_lease_owner_id
      AND job.lease_generation = p_lease_generation
      AND job.lease_until > v_now
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5691',
            MESSAGE = 'event gap reconciliation job is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_group_generation,v_job.authority_domain,
        v_job.cutover_generation
    );
    SELECT root.* INTO v_root FROM platform.dead_letters AS root
    WHERE v_job.payload_ref = 'deadletter:event-gap:' || root.id::text
      AND root.record_kind = 'root' AND root.source_kind = 'event'
      AND root.reason_class = 'sequence_gap'
      AND root.reason_code = 'EVENT_SEQUENCE_GAP'
      AND root.payload_canonical IS NOT NULL
      AND root.payload_redacted_at IS NULL
      AND root.source_sha256 = v_job.payload_sha256
      AND ROW(root.tenant_id,root.scope_kind,root.scope_id,
          root.authority_group_id,root.authority_group_generation,
          root.authority_domain,root.cutover_generation,root.writer_origin)
          = ROW(v_job.tenant_id,v_job.scope_kind,v_job.scope_id,
              v_job.authority_group_id,v_job.authority_group_generation,
              v_job.authority_domain,v_job.cutover_generation,'v2')
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5694',
            MESSAGE = 'event gap evidence is missing or redacted';
    END IF;
    v_gap := pg_catalog.convert_from(v_root.payload_canonical,'UTF8')::jsonb;
    SELECT event.* INTO v_event FROM platform.outbox AS event
    WHERE event.event_id = v_root.source_id
      AND event.payload_canonical IS NOT NULL
      AND event.payload_redacted_at IS NULL
      AND event.event_id::text = v_gap->'envelope'->>'event_id'
      AND event.event_sequence =
          (v_gap->'envelope'->>'event_sequence')::bigint
      AND event.payload_sha256 = v_gap->'envelope'->>'payload_sha256'
      AND pg_catalog.convert_from(event.payload_canonical,'UTF8')::jsonb =
          v_gap->'envelope'->'payload'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5693',
            MESSAGE = 'event gap source outbox is unavailable';
    END IF;
    IF platform.platform_async_outbox_authority_current_v1(v_event) THEN
        SELECT result.completed_at,result.requeued,result.action_id
        INTO completed_at,requeued,action_id
        FROM platform.platform_async_reconcile_event_gap_v1(
            p_job_id,p_attempt_id,p_lease_owner_id,p_lease_generation
        ) AS result;
        disposition := CASE WHEN requeued THEN 'requeued' ELSE 'already_applied' END;
        RETURN NEXT;
        RETURN;
    END IF;

    v_now := GREATEST(
        pg_catalog.clock_timestamp(),v_job.updated_at + interval '1 microsecond'
    );
    v_error_sha256 := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'EVENT_GAP_SOURCE_AUTHORITY_STALE','UTF8')),'hex');
    PERFORM platform.platform_async_record_event_gap_action_v1(
        v_job,v_root,v_gap,'repair_action',
        'automatic sequence-gap reconciliation terminated because source authority is stale',
        v_now
    );
    UPDATE platform.job_attempts AS attempt
    SET finished_at = v_now,outcome = 'cancelled',error_class = 'cancelled',
        error_code = 'EVENT_GAP_SOURCE_AUTHORITY_STALE',
        redacted_error_sha256 = v_error_sha256
    WHERE attempt.attempt_id = p_attempt_id AND attempt.job_id = v_job.id
      AND attempt.attempt = v_job.attempts
      AND attempt.lease_owner_id = p_lease_owner_id
      AND attempt.lease_generation = p_lease_generation
      AND attempt.finished_at IS NULL AND attempt.effect_started_at IS NULL;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5695',
            MESSAGE = 'event gap source-stale attempt completion failed';
    END IF;
    UPDATE platform.jobs AS job
    SET state = 'cancelled',lease_owner_id = NULL,lease_heartbeat_at = NULL,
        lease_until = NULL,last_error_class = 'cancelled',
        last_error_code = 'EVENT_GAP_SOURCE_AUTHORITY_STALE',
        last_error_sha256 = v_error_sha256,
        retain_until = v_now + interval '90 days',updated_at = v_now,
        completed_at = v_now
    WHERE job.id = v_job.id AND job.state = 'running'
      AND job.lease_owner_id = p_lease_owner_id
      AND job.lease_generation = p_lease_generation;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5696',
            MESSAGE = 'event gap source-stale job completion failed';
    END IF;
    completed_at := v_now;
    requeued := false;
    action_id := v_job.id;
    disposition := 'source_authority_stale';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_reconcile_event_gap_v2(
    uuid,uuid,uuid,bigint
) FROM PUBLIC;
