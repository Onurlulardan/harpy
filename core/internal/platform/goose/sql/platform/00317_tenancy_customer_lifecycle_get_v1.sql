-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_get_v1(
    in_environment_id uuid,
    in_target_id uuid
)
RETURNS TABLE (
    id uuid,
    name text,
    slug text,
    status text,
    lifecycle_revision bigint,
    settings_revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    offboarded_at timestamptz,
    purged_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_get_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_get_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_get_v1(uuid, uuid);
