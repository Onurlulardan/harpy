-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_offering_v2(
    p_authority_id uuid, p_operation_id uuid, p_offering_id uuid,
    p_version_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_offering_key text, p_name text,
    p_description text, p_sort_order integer, p_highlight_label text,
    p_color text, p_billing_interval text, p_is_public boolean,
    p_is_base boolean, p_trial_days integer, p_grace_days integer,
    p_expected_price_count smallint, p_expected_grant_count smallint,
    p_effective_start timestamptz, p_effective_end timestamptz,
    p_occurred_at timestamptz
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
    v_metadata jsonb;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    IF pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR pg_catalog.uuid_extract_version(p_offering_id) <> 7
       OR pg_catalog.uuid_extract_version(p_version_id) <> 7
       OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR p_name IS NULL OR pg_catalog.octet_length(p_name) NOT BETWEEN 1 AND 128
       OR pg_catalog.btrim(p_name) <> p_name
       OR p_description IS NULL OR pg_catalog.octet_length(p_description) > 2048
       OR p_highlight_label IS NULL OR pg_catalog.octet_length(p_highlight_label) > 50
       OR pg_catalog.btrim(p_highlight_label) <> p_highlight_label
       OR p_color IS NULL OR pg_catalog.octet_length(p_color) > 20
       OR pg_catalog.btrim(p_color) <> p_color
       OR p_billing_interval NOT IN ('monthly', 'annual')
       OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing offering create input is invalid';
    END IF;
    v_metadata := pg_catalog.jsonb_build_object(
        'name', p_name, 'description', p_description,
        'sort_order', p_sort_order, 'highlight_label', p_highlight_label,
        'color', p_color, 'billing_interval', p_billing_interval
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT request_sha256, aggregate_id, result_version_id,
           result_revision, result_version, result_lifecycle
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
    INSERT INTO platform.billing_offerings (
        id, authority_id, offering_key, lifecycle, created_at, retired_at, revision
    ) VALUES (
        p_offering_id, p_authority_id, p_offering_key, 'active', p_occurred_at, NULL, 1
    );
    INSERT INTO platform.billing_offering_versions (
        id, authority_id, offering_id, version, is_base, is_public,
        trial_days, grace_days, expected_price_count, expected_grant_count,
        effective_start, effective_end, presentation_metadata,
        presentation_sha256, publication_state, publication_operation_id, published_at
    ) VALUES (
        p_version_id, p_authority_id, p_offering_id, 1, p_is_base, p_is_public,
        p_trial_days, p_grace_days, p_expected_price_count, p_expected_grant_count,
        p_effective_start, p_effective_end, v_metadata,
        pg_catalog.sha256(pg_catalog.convert_to(v_metadata::text, 'UTF8')),
        'published', p_operation_id, p_occurred_at
    );
    INSERT INTO platform.billing_catalog_operations (
        operation_id, authority_id, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, operation_kind, result_version_id,
        result_revision, result_version, result_lifecycle, reason, occurred_at
    ) VALUES (
        p_operation_id, p_authority_id, p_actor_id, p_idempotency_key,
        p_request_sha256, 'offering', p_offering_id, 'create', p_version_id,
        1, 1, 'active', NULL, p_occurred_at
    );
    RETURN QUERY SELECT p_offering_id, p_version_id, 1::bigint,
        1::bigint, 'active'::text, false;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_create_offering_v2(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,text,integer,text,text,text,
    boolean,boolean,integer,integer,smallint,smallint,timestamptz,timestamptz,
    timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_offering_v2(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,text,integer,text,text,text,
    boolean,boolean,integer,integer,smallint,smallint,timestamptz,timestamptz,
    timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
