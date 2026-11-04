-- +goose Up
-- BELIRSIZ: feature enablement is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_feature_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_entitlement_key text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_metered_resource_feature_v1(
    uuid, uuid, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_feature_v1(
    uuid, uuid, text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_metered_resource_feature_v1(
    uuid, uuid, text
);
