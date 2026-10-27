-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_release_legal_hold_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_hold_id uuid,
    p_operation_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_reason text,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    hold_id uuid,
    tenant_id uuid,
    state text,
    placed_by uuid,
    place_operation_id uuid,
    reason text,
    ticket_reference text,
    placed_at timestamptz,
    released_by uuid,
    release_operation_id uuid,
    release_reason text,
    released_at timestamptz,
    revision bigint,
    replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored platform.billing_legal_holds%ROWTYPE;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF NOT EXISTS (
        SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = p_authority_id
          AND authority_row.tenant_id = p_tenant_id
    ) OR pg_catalog.uuid_extract_version(p_hold_id) <> 7
       OR pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR pg_catalog.octet_length(p_reason) NOT BETWEEN 1 AND 2048
       OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing legal hold release input is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'billing:legal-hold:' || p_tenant_id::text, 0
    ));
    SELECT * INTO stored
    FROM platform.billing_legal_holds AS legal_hold
    WHERE legal_hold.tenant_id = p_tenant_id AND legal_hold.id = p_hold_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'billing legal hold does not exist';
    END IF;
    IF stored.state IS NOT DISTINCT FROM 'released' THEN
        IF stored.release_idempotency_key IS DISTINCT FROM p_idempotency_key
           OR stored.release_request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5730',
                MESSAGE = 'billing legal hold release key was reused';
        END IF;
        RETURN QUERY SELECT stored.id, stored.tenant_id, stored.state,
            stored.placed_by, stored.place_operation_id, stored.reason,
            stored.ticket_reference, stored.placed_at, stored.released_by,
            stored.release_operation_id, stored.release_reason,
            stored.released_at, stored.revision, true;
        RETURN;
    END IF;
    UPDATE platform.billing_legal_holds AS legal_hold
       SET state = 'released', released_by = p_actor_id,
           release_operation_id = p_operation_id,
           release_idempotency_key = p_idempotency_key,
           release_request_sha256 = p_request_sha256,
           release_reason = p_reason, released_at = p_occurred_at,
           revision = legal_hold.revision + 1
     WHERE legal_hold.tenant_id = p_tenant_id AND legal_hold.id = p_hold_id
     RETURNING * INTO stored;
    RETURN QUERY SELECT stored.id, stored.tenant_id, stored.state,
        stored.placed_by, stored.place_operation_id, stored.reason,
        stored.ticket_reference, stored.placed_at, stored.released_by,
        stored.release_operation_id, stored.release_reason,
        stored.released_at, stored.revision, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_release_legal_hold_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_release_legal_hold_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
