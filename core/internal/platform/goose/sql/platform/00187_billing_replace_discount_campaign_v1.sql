-- +goose Up
-- BELIRSIZ: discount campaign replacement is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_replace_discount_campaign_v1(
    p_authority_id uuid,
    p_operation_id uuid,
    p_discount_id uuid,
    p_previous_version_id uuid,
    p_version_id uuid,
    p_campaign_id uuid,
    p_association_id_0 uuid,
    p_association_id_1 uuid,
    p_association_id_2 uuid,
    p_association_id_3 uuid,
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

REVOKE ALL ON FUNCTION platform.billing_replace_discount_campaign_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, uuid, uuid, uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_replace_discount_campaign_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, uuid, uuid, uuid, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_replace_discount_campaign_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, uuid, uuid, uuid, timestamptz
);
