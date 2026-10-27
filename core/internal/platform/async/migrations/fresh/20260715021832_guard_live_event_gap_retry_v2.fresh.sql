-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_dead_letter_chain_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    root_row platform.dead_letters%ROWTYPE;
    expected_sequence bigint;
    live_gap_job boolean;
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'dead_letters' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5640',
            MESSAGE = 'dead-letter chain trigger target is invalid';
    END IF;

    IF (
        (NEW.source_kind = 'job' AND
            NEW.consumer_name IS NULL AND NEW.consumer_version IS NULL) OR
        (NEW.source_kind = 'event' AND
            NEW.consumer_name IS NOT NULL AND NEW.consumer_version IS NOT NULL)
    ) IS NOT TRUE THEN
        RAISE EXCEPTION USING ERRCODE = 'P5641',
            MESSAGE = 'dead-letter source consumer shape is invalid';
    END IF;

    IF NEW.record_kind = 'root' THEN
        IF NEW.id IS DISTINCT FROM NEW.root_id OR
           NEW.entry_sequence IS DISTINCT FROM 0 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5642',
                MESSAGE = 'dead-letter root identity is invalid';
        END IF;
        RETURN NEW;
    END IF;

    IF (NEW.record_kind IN ('retry_action', 'repair_action')) IS NOT TRUE OR
       NEW.writer_origin IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5642',
            MESSAGE = 'dead-letter action identity is invalid';
    END IF;

    SELECT root.* INTO root_row
    FROM platform.dead_letters AS root
    WHERE root.id = NEW.root_id AND root.record_kind = 'root'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5643',
            MESSAGE = 'dead-letter root is missing';
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM platform.jobs AS job
        WHERE job.payload_ref = 'deadletter:event-gap:' || root_row.id::text
          AND job.payload_sha256 = root_row.payload_sha256
          AND job.kind = 'platform_async.event_gap_reconcile.' ||
              job.authority_domain
          AND job.state = 'running'
    ) INTO live_gap_job;
    IF NEW.record_kind = 'retry_action' AND (
        root_row.payload_canonical IS NULL OR
        root_row.payload_redacted_at IS NOT NULL OR
        (pg_catalog.clock_timestamp() >= root_row.full_payload_expires_at AND
            NOT (root_row.reason_class = 'sequence_gap' AND live_gap_job))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5644',
            MESSAGE = 'dead-letter payload cannot be retried';
    END IF;

    SELECT pg_catalog.max(entry.entry_sequence)::bigint + 1
    INTO expected_sequence
    FROM platform.dead_letters AS entry
    WHERE entry.root_id = NEW.root_id;
    IF NEW.entry_sequence::bigint IS DISTINCT FROM expected_sequence THEN
        RAISE EXCEPTION USING ERRCODE = 'P5645',
            MESSAGE = 'dead-letter action sequence is not contiguous';
    END IF;

    IF ROW(
        NEW.tenant_id, NEW.scope_kind, NEW.scope_id,
        NEW.authority_group_id, NEW.authority_group_generation,
        NEW.authority_domain, NEW.cutover_generation
    ) IS DISTINCT FROM ROW(
        root_row.tenant_id, root_row.scope_kind, root_row.scope_id,
        root_row.authority_group_id, root_row.authority_group_generation,
        root_row.authority_domain, root_row.cutover_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5646',
            MESSAGE = 'dead-letter action authority differs from root';
    END IF;

    IF ROW(
        NEW.source_kind, NEW.source_id, NEW.consumer_name, NEW.consumer_version,
        NEW.source_contract, NEW.source_version, NEW.source_sha256,
        NEW.reason_class, NEW.reason_code, NEW.redacted_error_sha256
    ) IS DISTINCT FROM ROW(
        root_row.source_kind, root_row.source_id,
        root_row.consumer_name, root_row.consumer_version,
        root_row.source_contract, root_row.source_version, root_row.source_sha256,
        root_row.reason_class, root_row.reason_code, root_row.redacted_error_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5647',
            MESSAGE = 'dead-letter action source differs from root';
    END IF;

    IF ROW(
        NEW.payload_sha256, NEW.redacted_envelope,
        NEW.full_payload_expires_at, NEW.metadata_expires_at
    ) IS DISTINCT FROM ROW(
        root_row.payload_sha256, root_row.redacted_envelope,
        root_row.full_payload_expires_at, root_row.metadata_expires_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5648',
            MESSAGE = 'dead-letter action evidence differs from root';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_dead_letter_chain_v2()
FROM PUBLIC;

CREATE OR REPLACE TRIGGER dead_letters_chain_insert_guard
BEFORE INSERT ON platform.dead_letters
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_dead_letter_chain_v2();
ALTER TABLE platform.dead_letters
    ENABLE ALWAYS TRIGGER dead_letters_chain_insert_guard;
