-- +goose Up
-- BELIRSIZ: revision, version, lifecycle, and replay are not determined by the caller.
CREATE FUNCTION platform.billing_create_campaign_version_v2(
    p_authority_id uuid,
    p_operation_id uuid,
    p_campaign_id uuid,
    p_version_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_expected_revision bigint,
    p_priority integer,
    p_discount_count smallint,
    p_starts_at timestamptz,
    p_ends_at timestamptz,
    p_presentation text,
    p_presentation_sha256 bytea,
    p_name text,
    p_description text,
    p_banner_text text,
    p_auto_apply boolean,
    p_available boolean,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    campaign_id uuid,
    version_id uuid,
    revision bigint,
    version bigint,
    lifecycle text,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_campaign_id,
        p_version_id,
        0::bigint,
        0::bigint,
        ''::text,
        false;
$$;

REVOKE ALL ON FUNCTION platform.billing_create_campaign_version_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, bigint, integer, smallint, timestamptz, timestamptz,
    text, bytea, text, text, text, boolean,
    boolean, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_campaign_version_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, bigint, integer, smallint, timestamptz, timestamptz,
    text, bytea, text, text, text, boolean,
    boolean, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_create_campaign_version_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, bigint, integer, smallint, timestamptz, timestamptz,
    text, bytea, text, text, text, boolean,
    boolean, timestamptz
);
