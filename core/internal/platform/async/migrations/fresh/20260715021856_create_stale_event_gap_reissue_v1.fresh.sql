-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_reissue_stale_event_gap_v1(
    p_job platform.jobs,
    p_root_id uuid,
    p_group_generation bigint,
    p_domain_generation bigint,
    p_now timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_old platform.dead_letters%ROWTYPE;
    v_new platform.dead_letters%ROWTYPE;
    v_prefix text;
BEGIN
    v_prefix := 'platform_async.event_gap_reconcile.';
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5680',
            MESSAGE = 'stale event gap reissue authority is invalid';
    END IF;
    IF pg_catalog.left(p_job.kind,pg_catalog.length(v_prefix)) <> v_prefix THEN
        RETURN p_job.payload_ref;
    END IF;
    IF p_job.kind IS DISTINCT FROM (v_prefix || p_job.authority_domain) OR
       p_root_id IS NULL OR p_job.payload_ref IS NULL OR
       pg_catalog.left(p_job.payload_ref,
           pg_catalog.length('deadletter:event-gap:')) <>
           'deadletter:event-gap:' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5681',
            MESSAGE = 'stale event gap source identity is invalid';
    END IF;

    SELECT root.* INTO v_old
    FROM platform.dead_letters AS root
    WHERE p_job.payload_ref = 'deadletter:event-gap:' || root.id::text
      AND root.record_kind = 'root' AND root.source_kind = 'event'
      AND root.reason_class = 'sequence_gap'
      AND root.reason_code = 'EVENT_SEQUENCE_GAP'
      AND root.payload_canonical IS NOT NULL
      AND root.payload_redacted_at IS NULL
      AND root.source_sha256 = p_job.payload_sha256
      AND ROW(root.tenant_id,root.scope_kind,root.scope_id,
          root.authority_group_id,root.authority_group_generation,
          root.authority_domain,root.cutover_generation,root.writer_origin)
          = ROW(p_job.tenant_id,p_job.scope_kind,p_job.scope_id,
              p_job.authority_group_id,p_job.authority_group_generation,
              p_job.authority_domain,p_job.cutover_generation,'v2')
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5682',
            MESSAGE = 'stale event gap evidence is unavailable';
    END IF;

    INSERT INTO platform.dead_letters (
        id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,source_kind,source_id,consumer_name,
        consumer_version,source_contract,source_version,source_sha256,
        reason_class,reason_code,redacted_error_sha256,payload_canonical,
        payload_sha256,redacted_envelope,full_payload_expires_at,
        metadata_expires_at,recorded_at
    ) VALUES (
        p_root_id,p_root_id,0,'root',v_old.tenant_id,v_old.scope_kind,
        v_old.scope_id,v_old.authority_group_id,p_group_generation,
        v_old.authority_domain,p_domain_generation,'v2',v_old.source_kind,
        v_old.source_id,v_old.consumer_name,v_old.consumer_version,
        v_old.source_contract,v_old.source_version,v_old.source_sha256,
        v_old.reason_class,v_old.reason_code,v_old.redacted_error_sha256,
        v_old.payload_canonical,v_old.payload_sha256,v_old.redacted_envelope,
        p_now + interval '30 days',p_now + interval '365 days',p_now
    ) ON CONFLICT (id) DO NOTHING;

    SELECT root.* INTO v_new
    FROM platform.dead_letters AS root
    WHERE ROW(root.id,root.root_id,root.entry_sequence,root.record_kind,
            root.tenant_id,root.scope_kind,root.scope_id,
            root.authority_group_id,root.authority_group_generation,
            root.authority_domain,root.cutover_generation,root.writer_origin,
            root.source_kind,root.source_id,root.consumer_name,
            root.consumer_version,root.source_contract,root.source_version,
            root.source_sha256,root.reason_class,root.reason_code,
            root.redacted_error_sha256,root.payload_canonical,
            root.payload_sha256,root.redacted_envelope,
            root.full_payload_expires_at,root.metadata_expires_at,
            root.recorded_at,root.payload_redacted_at)
          IS NOT DISTINCT FROM
          ROW(p_root_id,p_root_id,0,'root',v_old.tenant_id,v_old.scope_kind,
            v_old.scope_id,v_old.authority_group_id,p_group_generation,
            v_old.authority_domain,p_domain_generation,'v2',v_old.source_kind,
            v_old.source_id,v_old.consumer_name,v_old.consumer_version,
            v_old.source_contract,v_old.source_version,v_old.source_sha256,
            v_old.reason_class,v_old.reason_code,v_old.redacted_error_sha256,
            v_old.payload_canonical,v_old.payload_sha256,
            v_old.redacted_envelope,p_now + interval '30 days',
            p_now + interval '365 days',p_now,NULL);
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5683',
            MESSAGE = 'stale event gap replacement conflicts';
    END IF;
    RETURN 'deadletter:event-gap:' || p_root_id::text;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_reissue_stale_event_gap_v1(
    platform.jobs,uuid,bigint,bigint,timestamptz
) FROM PUBLIC;
