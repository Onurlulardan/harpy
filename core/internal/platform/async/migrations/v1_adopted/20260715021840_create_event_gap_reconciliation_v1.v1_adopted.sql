-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_reconcile_event_gap_v1(
    p_job_id uuid,
    p_attempt_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint
)
RETURNS TABLE(completed_at timestamptz, requeued boolean, action_id uuid)
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
    v_through bigint;
    v_duplicate boolean;
    v_action_kind text;
    v_action_reason text;
    v_count bigint;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5690',
            MESSAGE = 'event gap reconciliation authority is invalid';
    END IF;

    SELECT job.* INTO v_job
    FROM platform.jobs AS job
    WHERE job.id = p_job_id
      AND job.state = 'running'
      AND job.kind = 'platform_async.event_gap_reconcile.' ||
          job.authority_domain
      AND job.payload_revision = 1
      AND job.effect_class = 'database_transactional'
      AND job.effect_started_at IS NULL
      AND job.cancel_requested_at IS NULL
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

    SELECT root.* INTO v_root
    FROM platform.dead_letters AS root
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
    v_gap := pg_catalog.convert_from(v_root.payload_canonical, 'UTF8')::jsonb;

    SELECT event.* INTO v_event
    FROM platform.outbox AS event
    WHERE event.event_id = v_root.source_id
      AND event.payload_canonical IS NOT NULL
      AND event.payload_redacted_at IS NULL
      AND event.event_id::text = v_gap->'envelope'->>'event_id'
      AND event.event_sequence =
          (v_gap->'envelope'->>'event_sequence')::bigint
      AND event.payload_sha256 = v_gap->'envelope'->>'payload_sha256'
      AND pg_catalog.convert_from(event.payload_canonical, 'UTF8')::jsonb =
          v_gap->'envelope'->'payload'
    FOR UPDATE;
    IF NOT FOUND OR v_event.parked_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5693',
            MESSAGE = 'event gap source outbox is unavailable';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        pg_catalog.concat_ws(E'\x1f',
            'platform_async_inbox_v1',v_root.consumer_name,
            v_root.consumer_version::text,v_root.scope_kind,
            v_root.scope_id::text,v_event.aggregate_type,
            v_event.aggregate_id::text
        ), 0
    ));
    SELECT
        COALESCE(pg_catalog.max(mark.through_sequence), 0),
        COALESCE(pg_catalog.bool_or(
            mark.record_kind = 'event_receipt' AND
            mark.event_id = v_root.source_id
        ), false)
    INTO v_through,v_duplicate
    FROM platform.inbox AS mark
    WHERE mark.consumer_name = v_root.consumer_name
      AND mark.consumer_version = v_root.consumer_version
      AND mark.scope_kind = v_root.scope_kind
      AND mark.scope_id = v_root.scope_id
      AND mark.aggregate_type = v_event.aggregate_type
      AND mark.aggregate_id = v_event.aggregate_id;

    IF v_duplicate OR v_through >= v_event.event_sequence THEN
        v_action_kind := 'repair_action';
        v_action_reason := 'automatic sequence-gap reconciliation found durable inbox evidence';
    ELSIF v_through = v_event.event_sequence - 1 AND
          v_event.delivery_state IN ('pending','publishing') THEN
        v_action_kind := 'repair_action';
        v_action_reason := 'automatic sequence-gap replay was already pending';
    ELSIF v_through = v_event.event_sequence - 1 AND
          v_event.delivery_state = 'delivered' THEN
        PERFORM platform.platform_async_assert_current_authority_v1(
            v_event.scope_kind,v_event.scope_id,v_event.authority_group_id,
            v_event.authority_group_generation,v_event.authority_domain,
            v_event.cutover_generation
        );
        v_action_kind := 'retry_action';
        v_action_reason := 'automatic sequence-gap replay after contiguous watermark';
    ELSE
        RAISE EXCEPTION USING ERRCODE = 'P5692',
            MESSAGE = 'event gap predecessor watermark is not ready';
    END IF;

    PERFORM platform.platform_async_record_event_gap_action_v1(
        v_job,v_root,v_gap,v_action_kind,v_action_reason,v_now
    );

    requeued := v_action_kind = 'retry_action';
    IF requeued THEN
        UPDATE platform.outbox AS event
        SET delivery_state = 'pending',next_attempt_at = v_now,
            confirmed_at = NULL,delivered_at = NULL,confirm_sha256 = NULL,
            last_error_code = NULL,last_error_sha256 = NULL
        WHERE event.event_id = v_event.event_id;
        GET DIAGNOSTICS v_count = ROW_COUNT;
        IF v_count <> 1 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5693',
                MESSAGE = 'event gap source outbox was not reopened';
        END IF;
    END IF;

    completed_at := platform.platform_async_complete_event_gap_job_v1(
        v_job,p_attempt_id,p_lease_owner_id,p_lease_generation
    );
    action_id := v_job.id;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_reconcile_event_gap_v1(
    uuid,uuid,uuid,bigint
) FROM PUBLIC;
