-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_campaign_v1(
    p_authority_id uuid, p_operation_id uuid, p_campaign_id uuid,
    p_version_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_campaign_key text, p_priority integer,
    p_expected_discount_count smallint, p_starts_at timestamptz,
    p_ends_at timestamptz, p_presentation_metadata jsonb,
    p_presentation_sha256 bytea, p_occurred_at timestamptz
)
RETURNS TABLE (
    result_campaign_id uuid, result_version_id uuid,
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
    stored_aggregate_kind text;
    stored_operation_kind text;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT request_sha256, aggregate_id,
           billing_catalog_operations.result_version_id,
           billing_catalog_operations.result_revision,
           billing_catalog_operations.result_version,
           billing_catalog_operations.result_lifecycle,
           aggregate_kind, operation_kind
      INTO stored_request_sha256, stored_aggregate_id, stored_version_id,
           stored_revision, stored_version, stored_lifecycle,
           stored_aggregate_kind, stored_operation_kind
    FROM platform.billing_catalog_operations
    WHERE authority_id = p_authority_id AND actor_id = p_actor_id
      AND idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF stored_request_sha256 IS DISTINCT FROM p_request_sha256
           OR stored_aggregate_kind IS DISTINCT FROM 'campaign'
           OR stored_operation_kind IS DISTINCT FROM 'create' THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing catalog idempotency key was reused';
        END IF;
        RETURN QUERY SELECT stored_aggregate_id, stored_version_id,
            stored_revision, stored_version, stored_lifecycle, true;
        RETURN;
    END IF;
    INSERT INTO platform.billing_campaigns (
        id, authority_id, campaign_key, lifecycle, created_at, retired_at, revision
    ) VALUES (
        p_campaign_id, p_authority_id, p_campaign_key,
        'active', p_occurred_at, NULL, 1
    );
    INSERT INTO platform.billing_campaign_versions (
        id, authority_id, campaign_id, version, priority,
        expected_discount_count, starts_at, ends_at, presentation_metadata,
        presentation_sha256, publication_state, publication_operation_id,
        published_at
    ) VALUES (
        p_version_id, p_authority_id, p_campaign_id, 1, p_priority,
        p_expected_discount_count, p_starts_at, p_ends_at,
        p_presentation_metadata, p_presentation_sha256,
        'published', p_operation_id, p_occurred_at
    );
    INSERT INTO platform.billing_catalog_operations (
        operation_id, authority_id, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, operation_kind, result_version_id,
        result_revision, result_version, result_lifecycle, reason, occurred_at
    ) VALUES (
        p_operation_id, p_authority_id, p_actor_id, p_idempotency_key,
        p_request_sha256, 'campaign', p_campaign_id, 'create',
        p_version_id, 1, 1, 'active', NULL, p_occurred_at
    );
    RETURN QUERY SELECT p_campaign_id, p_version_id, 1::bigint,
        1::bigint, 'active'::text, false;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_create_campaign_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,integer,smallint,
    timestamptz,timestamptz,jsonb,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_campaign_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,integer,smallint,
    timestamptz,timestamptz,jsonb,bytea,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
