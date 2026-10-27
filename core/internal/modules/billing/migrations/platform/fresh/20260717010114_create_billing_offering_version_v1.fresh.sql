-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_offering_version_v1(
    p_authority_id uuid, p_operation_id uuid, p_offering_id uuid,
    p_version_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_expected_revision bigint, p_name text,
    p_is_public boolean, p_is_base boolean, p_trial_days integer,
    p_grace_days integer, p_expected_price_count smallint,
    p_expected_grant_count smallint, p_effective_start timestamptz,
    p_effective_end timestamptz, p_occurred_at timestamptz
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
    stored_version_id uuid;
    stored_revision bigint;
    stored_version bigint;
    stored_lifecycle text;
    updated_revision bigint;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT request_sha256, aggregate_id,
           billing_catalog_operations.result_version_id,
           billing_catalog_operations.result_revision,
           billing_catalog_operations.result_version,
           billing_catalog_operations.result_lifecycle
      INTO stored_request_sha256, stored_aggregate_id, stored_version_id,
           stored_revision, stored_version, stored_lifecycle
    FROM platform.billing_catalog_operations
    WHERE authority_id = p_authority_id AND actor_id = p_actor_id
      AND idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF stored_request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing catalog idempotency key was reused';
        END IF;
        RETURN QUERY SELECT stored_aggregate_id, stored_version_id,
            stored_revision, stored_version, stored_lifecycle, true;
        RETURN;
    END IF;
    UPDATE platform.billing_offerings
    SET revision = revision + 1
    WHERE authority_id = p_authority_id AND id = p_offering_id
      AND lifecycle = 'active' AND revision = p_expected_revision
    RETURNING revision INTO updated_revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing offering revision is stale';
    END IF;
    INSERT INTO platform.billing_offering_versions (
        id, authority_id, offering_id, version, is_base, is_public,
        trial_days, grace_days, expected_price_count, expected_grant_count,
        effective_start, effective_end, presentation_metadata,
        presentation_sha256, publication_state,
        publication_operation_id, published_at
    ) VALUES (
        p_version_id, p_authority_id, p_offering_id, updated_revision,
        p_is_base, p_is_public, p_trial_days, p_grace_days,
        p_expected_price_count, p_expected_grant_count,
        p_effective_start, p_effective_end,
        pg_catalog.jsonb_build_object('name', p_name),
        pg_catalog.sha256(pg_catalog.convert_to(
            pg_catalog.jsonb_build_object('name', p_name)::text, 'UTF8'
        )), 'published', p_operation_id, p_occurred_at
    );
    INSERT INTO platform.billing_catalog_operations (
        operation_id, authority_id, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, operation_kind, result_version_id,
        result_revision, result_version, result_lifecycle, reason, occurred_at
    ) VALUES (
        p_operation_id, p_authority_id, p_actor_id, p_idempotency_key,
        p_request_sha256, 'offering', p_offering_id, 'create_version',
        p_version_id, updated_revision, updated_revision,
        'active', NULL, p_occurred_at
    );
    RETURN QUERY SELECT p_offering_id, p_version_id, updated_revision,
        updated_revision, 'active'::text, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_create_offering_version_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,bigint,text,boolean,boolean,
    integer,integer,smallint,smallint,timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_offering_version_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,bigint,text,boolean,boolean,
    integer,integer,smallint,smallint,timestamptz,timestamptz,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
