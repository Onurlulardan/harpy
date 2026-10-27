-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_park_stale_outbox_v1(
    p_event_id uuid,
    p_dead_letter_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_event platform.outbox%ROWTYPE;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_error_sha256 text;
BEGIN
    IF SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' OR
       p_event_id IS NULL OR p_dead_letter_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5653',
            MESSAGE = 'stale outbox park input is invalid';
    END IF;

    SELECT * INTO v_event
    FROM platform.outbox
    WHERE event_id OPERATOR(pg_catalog.=) p_event_id;
    IF NOT FOUND OR v_event.parked_at IS NOT NULL THEN
        RETURN false;
    END IF;

    SELECT generation INTO v_group_generation
    FROM platform.cutover_groups
    WHERE scope_kind OPERATOR(pg_catalog.=) v_event.scope_kind
      AND scope_id OPERATOR(pg_catalog.=) v_event.scope_id
      AND group_id OPERATOR(pg_catalog.=) v_event.authority_group_id
      AND state OPERATOR(pg_catalog.=) 'v2'
      AND writer OPERATOR(pg_catalog.=) 'v2'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5654',
            MESSAGE = 'current outbox group authority is unavailable';
    END IF;

    SELECT generation INTO v_domain_generation
    FROM platform.domain_cutovers
    WHERE scope_kind OPERATOR(pg_catalog.=) v_event.scope_kind
      AND scope_id OPERATOR(pg_catalog.=) v_event.scope_id
      AND domain OPERATOR(pg_catalog.=) v_event.authority_domain
      AND state OPERATOR(pg_catalog.=) 'v2'
      AND writer OPERATOR(pg_catalog.=) 'v2'
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind OPERATOR(pg_catalog.=) v_event.scope_kind
          AND scope_id OPERATOR(pg_catalog.=) v_event.scope_id
          AND group_id OPERATOR(pg_catalog.=) v_event.authority_group_id
          AND domain OPERATOR(pg_catalog.=) v_event.authority_domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5655',
            MESSAGE = 'current outbox domain authority is unavailable';
    END IF;

    SELECT * INTO v_event
    FROM platform.outbox
    WHERE event_id OPERATOR(pg_catalog.=) p_event_id
    FOR UPDATE;
    IF v_event.parked_at IS NOT NULL OR
       NOT (
           v_event.delivery_state OPERATOR(pg_catalog.=) 'pending' OR
           (v_event.delivery_state OPERATOR(pg_catalog.=) 'publishing' AND
            v_event.lease_until OPERATOR(pg_catalog.<=) v_now)
       ) THEN
        RETURN false;
    END IF;
    IF v_group_generation < v_event.authority_group_generation OR
       v_domain_generation < v_event.cutover_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5656',
            MESSAGE = 'outbox authority generation moved backwards';
    END IF;
    IF v_group_generation = v_event.authority_group_generation AND
       v_domain_generation = v_event.cutover_generation THEN
        RETURN false;
    END IF;

    UPDATE platform.outbox
    SET delivery_state = 'pending',
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        parked_at = v_now,
        parked_code = 'OUTBOX_AUTHORITY_STALE',
        parked_current_group_generation = v_group_generation,
        parked_current_cutover_generation = v_domain_generation,
        updated_at = v_now
    WHERE event_id OPERATOR(pg_catalog.=) v_event.event_id;

    v_error_sha256 := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(
            'outbox-authority-stale.v1' || pg_catalog.chr(10) ||
            v_event.event_id::text || pg_catalog.chr(10) ||
            v_event.authority_group_generation::text || pg_catalog.chr(10) ||
            v_event.cutover_generation::text || pg_catalog.chr(10) ||
            v_group_generation::text || pg_catalog.chr(10) ||
            v_domain_generation::text,
            'UTF8'
        )
    ), 'hex');

    INSERT INTO platform.dead_letters (
        id, root_id, entry_sequence, record_kind,
        tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        source_kind, source_id, consumer_name, consumer_version,
        source_contract, source_version, source_sha256,
        reason_class, reason_code, redacted_error_sha256,
        payload_canonical, payload_sha256, redacted_envelope,
        full_payload_expires_at, metadata_expires_at, recorded_at
    ) VALUES (
        p_dead_letter_id, p_dead_letter_id, 0, 'root',
        v_event.tenant_id, v_event.scope_kind, v_event.scope_id,
        v_event.authority_group_id, v_group_generation,
        v_event.authority_domain, v_domain_generation, 'v2',
        'event', v_event.event_id, 'platform.outbox.publisher', 1,
        v_event.envelope_contract_version, v_event.event_version,
        v_event.payload_sha256,
        'stale_generation', 'OUTBOX_AUTHORITY_STALE', v_error_sha256,
        v_event.payload_canonical, v_event.payload_sha256,
        pg_catalog.jsonb_build_object(
            'contract', 'outbox-authority-stale.v1',
            'event_id', v_event.event_id,
            'event_type', v_event.event_type,
            'stored_group_generation', v_event.authority_group_generation,
            'stored_domain_generation', v_event.cutover_generation,
            'current_group_generation', v_group_generation,
            'current_domain_generation', v_domain_generation
        ),
        v_now + interval '30 days', v_now + interval '365 days', v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_park_stale_outbox_v1(
    uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_park_stale_outbox_v1(
    uuid,uuid
) TO forgecrud_runtime;

