-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(
    in_tenant_id uuid,
    in_artifact_id uuid,
    in_outcome text
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

REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(uuid, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(uuid, uuid, text);
