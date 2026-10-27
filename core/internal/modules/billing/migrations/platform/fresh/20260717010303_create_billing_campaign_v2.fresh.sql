-- +goose Up
-- Migration version: 20260717010303 UTC
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_campaign_v2(
    p_authority_id uuid, p_operation_id uuid, p_campaign_id uuid,
    p_version_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_campaign_key text, p_priority integer,
    p_expected_discount_count smallint, p_starts_at timestamptz,
    p_ends_at timestamptz, p_presentation_metadata jsonb,
    p_presentation_sha256 bytea, p_name text, p_description text,
    p_banner_text text, p_auto_apply boolean, p_available boolean,
    p_occurred_at timestamptz
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
    stored platform.billing_catalog_operations%ROWTYPE;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT operation.* INTO stored
    FROM platform.billing_catalog_operations AS operation
    WHERE operation.authority_id = p_authority_id
      AND operation.actor_id = p_actor_id
      AND operation.idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF stored.request_sha256 IS DISTINCT FROM p_request_sha256
           OR stored.aggregate_kind IS DISTINCT FROM 'campaign'
           OR stored.operation_kind IS DISTINCT FROM 'create' THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing catalog idempotency key was reused';
        END IF;
        RETURN QUERY SELECT stored.aggregate_id, stored.result_version_id,
            stored.result_revision, stored.result_version,
            stored.result_lifecycle, true;
        RETURN;
    END IF;
    INSERT INTO platform.billing_campaigns (
        id, authority_id, campaign_key, lifecycle,
        created_at, retired_at, revision
    ) VALUES (
        p_campaign_id, p_authority_id, p_campaign_key,
        'active', p_occurred_at, NULL, 1
    );
    INSERT INTO platform.billing_campaign_versions (
        id, authority_id, campaign_id, version, priority,
        expected_discount_count, starts_at, ends_at,
        presentation_metadata, presentation_sha256, publication_state,
        publication_operation_id, published_at, presentation_name,
        presentation_description, presentation_banner_text, auto_apply,
        available, superseded_version_id
    ) VALUES (
        p_version_id, p_authority_id, p_campaign_id, 1, p_priority,
        p_expected_discount_count, p_starts_at, p_ends_at,
        p_presentation_metadata, p_presentation_sha256, 'published',
        p_operation_id, p_occurred_at, p_name, p_description,
        p_banner_text, p_auto_apply, p_available, NULL
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
REVOKE ALL ON FUNCTION platform.billing_create_campaign_v2(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,integer,smallint,
    timestamptz,timestamptz,jsonb,bytea,text,text,text,boolean,boolean,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_campaign_v2(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,integer,smallint,
    timestamptz,timestamptz,jsonb,bytea,text,text,text,boolean,boolean,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
