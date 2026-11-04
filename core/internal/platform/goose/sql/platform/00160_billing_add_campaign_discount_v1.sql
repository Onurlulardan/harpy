-- +goose Up
-- BELIRSIZ: campaign discount association is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_campaign_discount_v1(
    p_authority_id uuid,
    p_version_id uuid,
    p_discount_version_id uuid,
    p_occurred_at timestamptz
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT CAST(NULL AS pg_catalog.void);
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_add_campaign_discount_v1(
    uuid, uuid, uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_campaign_discount_v1(
    uuid, uuid, uuid, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_add_campaign_discount_v1(
    uuid, uuid, uuid, timestamptz
);
