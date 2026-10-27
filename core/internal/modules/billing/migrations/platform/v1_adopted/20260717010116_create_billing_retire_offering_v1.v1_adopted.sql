-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_retire_offering_v1(
    p_authority_id uuid, p_operation_id uuid, p_offering_id uuid,
    p_actor_id uuid, p_idempotency_key text, p_request_sha256 bytea,
    p_expected_revision bigint, p_reason text, p_occurred_at timestamptz
)
RETURNS TABLE (
    result_offering_id uuid, result_version_id uuid,
    result_revision bigint, result_version bigint,
    result_lifecycle text, replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored_request_sha256 bytea;
    stored_aggregate_id uuid;
    stored_revision bigint;
    stored_lifecycle text;
    updated_revision bigint;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT request_sha256, aggregate_id,
           billing_catalog_operations.result_revision,
           billing_catalog_operations.result_lifecycle
      INTO stored_request_sha256, stored_aggregate_id,
           stored_revision, stored_lifecycle
    FROM platform.billing_catalog_operations
    WHERE authority_id = p_authority_id AND actor_id = p_actor_id
      AND idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF stored_request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing catalog idempotency key was reused';
        END IF;
        RETURN QUERY SELECT stored_aggregate_id, NULL::uuid,
            stored_revision, NULL::bigint, stored_lifecycle, true;
        RETURN;
    END IF;
    UPDATE platform.billing_offerings
    SET lifecycle = 'retired', retired_at = p_occurred_at,
        revision = revision + 1
    WHERE authority_id = p_authority_id AND id = p_offering_id
      AND lifecycle = 'active' AND revision = p_expected_revision
    RETURNING revision INTO updated_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing offering revision is stale';
    END IF;
    INSERT INTO platform.billing_catalog_operations (
        operation_id, authority_id, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, operation_kind, result_version_id,
        result_revision, result_version, result_lifecycle, reason, occurred_at
    ) VALUES (
        p_operation_id, p_authority_id, p_actor_id, p_idempotency_key,
        p_request_sha256, 'offering', p_offering_id, 'retire', NULL,
        updated_revision, NULL, 'retired', p_reason, p_occurred_at
    );
    RETURN QUERY SELECT p_offering_id, NULL::uuid, updated_revision,
        NULL::bigint, 'retired'::text, false;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_retire_offering_v1(
    uuid,uuid,uuid,uuid,text,bytea,bigint,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_retire_offering_v1(
    uuid,uuid,uuid,uuid,text,bytea,bigint,text,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
