-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_list_v1(
    in_environment_id uuid,
    in_limit integer,
    in_after uuid
)
RETURNS TABLE (
    document bytea,
    next_cursor uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        NULL::uuid
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_list_v1(uuid, integer, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_list_v1(uuid, integer, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_list_v1(uuid, integer, uuid);
