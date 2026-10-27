-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_terminal_job_dead_letter_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_attempt platform.job_attempts%ROWTYPE;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_root_digest text;
    v_root_id uuid;
    v_payload bytea;
    v_payload_sha256 text;
    v_error_sha256 text;
    v_reason_class text;
    v_redacted_envelope jsonb;
    v_inserted integer;
    v_existing platform.dead_letters%ROWTYPE;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'jobs' OR
       pg_catalog.pg_trigger_depth() IS DISTINCT FROM 1 OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       OLD.state IS NOT DISTINCT FROM 'dead' OR NEW.state IS DISTINCT FROM 'dead' OR
       NEW.completed_at IS NULL OR NEW.completed_at IS DISTINCT FROM NEW.updated_at OR
       NEW.payload_ref IS NULL OR NEW.last_error_code IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5680',
            MESSAGE = 'terminal job dead-letter invocation is invalid';
    END IF;

    SELECT attempt.* INTO v_attempt
    FROM platform.job_attempts AS attempt
    WHERE attempt.job_id = NEW.id AND attempt.attempt = NEW.attempts
    FOR SHARE;
    IF NOT FOUND OR v_attempt.finished_at IS NULL OR v_attempt.outcome <> 'dead' OR
       v_attempt.error_class IS DISTINCT FROM NEW.last_error_class OR
       v_attempt.error_code IS DISTINCT FROM NEW.last_error_code OR
       v_attempt.redacted_error_sha256 IS DISTINCT FROM NEW.last_error_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5681',
            MESSAGE = 'terminal job dead-letter lacks exact attempt evidence';
    END IF;

    SELECT authority.generation INTO v_group_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = NEW.scope_kind AND authority.scope_id = NEW.scope_id
      AND authority.group_id = NEW.authority_group_id
      AND authority.state = 'v2' AND authority.writer = 'v2'
    FOR SHARE;
    SELECT authority.generation INTO v_domain_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = NEW.scope_kind AND authority.scope_id = NEW.scope_id
      AND authority.domain = NEW.authority_domain
      AND authority.state = 'v2' AND authority.writer = 'v2'
    FOR SHARE;
    IF v_group_generation IS NULL OR v_domain_generation IS NULL OR
       v_group_generation < NEW.authority_group_generation OR
       v_domain_generation < NEW.cutover_generation OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = NEW.scope_kind AND member.scope_id = NEW.scope_id
             AND member.group_id = NEW.authority_group_id
             AND member.domain = NEW.authority_domain
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5682',
            MESSAGE = 'terminal job dead-letter authority is unavailable';
    END IF;

    v_reason_class := CASE
        WHEN NEW.last_error_code = 'JOB_OUTCOME_UNKNOWN' THEN 'outcome_unknown'
        WHEN NEW.last_error_class = 'permanent' THEN 'permanent'
        ELSE 'delivery_limit'
    END;
    v_error_sha256 := COALESCE(
        NEW.last_error_sha256,
        pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.convert_to(NEW.last_error_code, 'UTF8')
        ), 'hex')
    );
    v_payload := pg_catalog.convert_to(
        '{"attempt":' || NEW.attempts::text ||
        ',"contract_version":"forgecrud.job-terminal-reference.v1","job_id":' ||
        pg_catalog.to_json(NEW.id::text)::text || ',"kind":' ||
        pg_catalog.to_json(NEW.kind)::text || ',"payload_reference":' ||
        pg_catalog.to_json(NEW.payload_ref)::text || ',"payload_revision":' ||
        NEW.payload_revision::text || ',"payload_sha256":' ||
        pg_catalog.to_json(NEW.payload_sha256)::text || ',"reason_code":' ||
        pg_catalog.to_json(NEW.last_error_code)::text || '}', 'UTF8'
    );
    v_payload_sha256 := pg_catalog.encode(pg_catalog.sha256(v_payload), 'hex');
    v_redacted_envelope := pg_catalog.jsonb_build_object(
        'attempt', NEW.attempts,
        'contract_version', 'forgecrud.job-terminal-reference.v1',
        'job_id', NEW.id, 'kind', NEW.kind,
        'payload_revision', NEW.payload_revision,
        'payload_sha256', NEW.payload_sha256,
        'reason_code', NEW.last_error_code
    );
    v_root_digest := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'forgecrud.job-terminal-dead-letter.v1' || pg_catalog.chr(10) ||
        'job_id=' || NEW.id::text || pg_catalog.chr(10) ||
        'attempt=' || NEW.attempts::text || pg_catalog.chr(10) ||
        'reason=' || NEW.last_error_code, 'UTF8'
    )), 'hex');
    v_root_id := (
        pg_catalog.substr(v_root_digest, 1, 8) || '-' ||
        pg_catalog.substr(v_root_digest, 9, 4) || '-8' ||
        pg_catalog.substr(v_root_digest, 14, 3) || '-' ||
        pg_catalog.substr('89ab', (
            (pg_catalog.get_byte(pg_catalog.decode(v_root_digest, 'hex'), 8) >> 4) & 3
        ) + 1, 1) ||
        pg_catalog.substr(v_root_digest, 18, 3) || '-' ||
        pg_catalog.substr(v_root_digest, 21, 12)
    )::uuid;

    INSERT INTO platform.dead_letters (
        id, root_id, entry_sequence, record_kind, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation, authority_domain,
        cutover_generation, writer_origin, source_kind, source_id, source_contract,
        source_version, source_sha256, reason_class, reason_code,
        redacted_error_sha256, payload_canonical, payload_sha256,
        redacted_envelope, full_payload_expires_at, metadata_expires_at, recorded_at
    ) VALUES (
        v_root_id, v_root_id, 0, 'root', NEW.tenant_id, NEW.scope_kind, NEW.scope_id,
        NEW.authority_group_id, v_group_generation, NEW.authority_domain,
        v_domain_generation, 'v2', 'job', NEW.id, NEW.kind,
        NEW.payload_revision, NEW.payload_sha256, v_reason_class, NEW.last_error_code,
        v_error_sha256, v_payload, v_payload_sha256,
        v_redacted_envelope,
        NEW.completed_at + interval '30 days',
        NEW.completed_at + interval '365 days', NEW.completed_at
    ) ON CONFLICT DO NOTHING;
    GET DIAGNOSTICS v_inserted = ROW_COUNT;
    IF v_inserted = 1 THEN
        RETURN NEW;
    END IF;

    SELECT root.* INTO v_existing
    FROM platform.dead_letters AS root
    WHERE root.source_kind = 'job' AND root.source_id = NEW.id
      AND root.record_kind = 'root'
      AND root.reason_class IN ('permanent','delivery_limit','outcome_unknown')
    FOR SHARE;
    IF NOT FOUND OR v_existing.id IS DISTINCT FROM v_root_id OR
       v_existing.tenant_id IS DISTINCT FROM NEW.tenant_id OR
       v_existing.scope_kind IS DISTINCT FROM NEW.scope_kind OR
       v_existing.scope_id IS DISTINCT FROM NEW.scope_id OR
       v_existing.authority_group_id IS DISTINCT FROM NEW.authority_group_id OR
       v_existing.authority_group_generation IS DISTINCT FROM v_group_generation OR
       v_existing.authority_domain IS DISTINCT FROM NEW.authority_domain OR
       v_existing.cutover_generation IS DISTINCT FROM v_domain_generation OR
       v_existing.writer_origin IS DISTINCT FROM 'v2' OR
       v_existing.source_contract IS DISTINCT FROM NEW.kind OR
       v_existing.source_version IS DISTINCT FROM NEW.payload_revision OR
       v_existing.reason_class IS DISTINCT FROM v_reason_class OR
       v_existing.reason_code IS DISTINCT FROM NEW.last_error_code OR
       v_existing.source_sha256 IS DISTINCT FROM NEW.payload_sha256 OR
       v_existing.redacted_error_sha256 IS DISTINCT FROM v_error_sha256 OR
       v_existing.payload_canonical IS DISTINCT FROM v_payload OR
       v_existing.payload_sha256 IS DISTINCT FROM v_payload_sha256 OR
       v_existing.redacted_envelope IS DISTINCT FROM v_redacted_envelope OR
       v_existing.recorded_at IS DISTINCT FROM NEW.completed_at OR
       v_existing.full_payload_expires_at IS DISTINCT FROM
           NEW.completed_at + interval '30 days' OR
       v_existing.metadata_expires_at IS DISTINCT FROM
           NEW.completed_at + interval '365 days' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5683',
            MESSAGE = 'terminal job dead-letter evidence conflicts';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_record_terminal_job_dead_letter_v1()
FROM PUBLIC;
