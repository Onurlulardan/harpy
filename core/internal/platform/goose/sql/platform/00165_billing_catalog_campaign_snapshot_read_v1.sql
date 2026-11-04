-- +goose Up
-- BELIRSIZ: campaign snapshot rows are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_catalog_campaign_snapshot_read_v1()
RETURNS TABLE (
    campaign_id uuid,
    legacy_id uuid,
    created_by uuid,
    campaign_key text,
    revision bigint,
    version_id uuid,
    version bigint,
    lifecycle text,
    publication_state text,
    name text,
    description text,
    banner_text text,
    auto_apply boolean,
    available boolean,
    priority integer,
    starts_at timestamptz,
    ends_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz,
    authority_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::bigint,
        NULL::uuid,
        NULL::bigint,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::boolean,
        NULL::integer,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::uuid
    WHERE false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_catalog_campaign_snapshot_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_catalog_campaign_snapshot_read_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_catalog_campaign_snapshot_read_v1();
