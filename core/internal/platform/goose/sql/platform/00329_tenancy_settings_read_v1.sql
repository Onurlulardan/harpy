-- +goose Up
CREATE FUNCTION platform.tenancy_settings_read_v1(
    in_tenant_id uuid
)
RETURNS TABLE (
    id uuid,
    name text,
    slug text,
    status text,
    settings_revision bigint,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp()
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_settings_read_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_settings_read_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_settings_read_v1(uuid);
