-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_event_gap_action_v1(
    p_job platform.jobs,
    p_root platform.dead_letters,
    p_gap jsonb,
    p_kind text,
    p_reason text,
    p_now timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_count bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_kind NOT IN ('retry_action','repair_action') OR
       p_job.payload_ref IS DISTINCT FROM
           'deadletter:event-gap:' || p_root.id::text OR
       p_root.reason_class IS DISTINCT FROM 'sequence_gap' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5697',
            MESSAGE = 'event gap action evidence is invalid';
    END IF;
    INSERT INTO platform.dead_letters (
        id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,source_kind,source_id,consumer_name,
        consumer_version,source_contract,source_version,source_sha256,
        reason_class,reason_code,redacted_error_sha256,payload_sha256,
        redacted_envelope,action_reason,action_ticket,action_actor_id,
        action_causation_id,full_payload_expires_at,metadata_expires_at,recorded_at
    ) SELECT
        p_job.id,p_root.id,pg_catalog.max(entry.entry_sequence) + 1,
        p_kind,p_root.tenant_id,p_root.scope_kind,p_root.scope_id,
        p_root.authority_group_id,p_root.authority_group_generation,
        p_root.authority_domain,p_root.cutover_generation,'v2',
        p_root.source_kind,p_root.source_id,p_root.consumer_name,
        p_root.consumer_version,p_root.source_contract,p_root.source_version,
        p_root.source_sha256,p_root.reason_class,p_root.reason_code,
        p_root.redacted_error_sha256,p_root.payload_sha256,
        p_root.redacted_envelope,p_reason,
        'event-gap-reconcile:' || p_job.id::text,
        (p_gap->'envelope'->>'actor_id')::uuid,p_job.id,
        p_root.full_payload_expires_at,p_root.metadata_expires_at,p_now
    FROM platform.dead_letters AS entry
    WHERE entry.root_id = p_root.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5697',
            MESSAGE = 'event gap action evidence was not recorded';
    END IF;
    RETURN p_job.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_record_event_gap_action_v1(
    platform.jobs,platform.dead_letters,jsonb,text,text,timestamptz
) FROM PUBLIC;

