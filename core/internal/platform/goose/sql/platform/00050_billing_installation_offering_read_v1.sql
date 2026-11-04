-- +goose Up
-- BELIRSIZ: installation offering identity is not determined by the caller.
CREATE FUNCTION platform.billing_installation_offering_read_v1(
    p_environment_id uuid,
    p_deployment_mode text
)
RETURNS TABLE (
    offering_version_id uuid,
    revision bigint,
    offering_sha256 text,
    effective_at timestamptz,
    source_sequence bigint,
    source_sha256 text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        pg_catalog.clock_timestamp(),
        NULL::bigint,
        NULL::text;
$$;

REVOKE ALL ON FUNCTION platform.billing_installation_offering_read_v1(
    uuid, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_offering_read_v1(
    uuid, text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_installation_offering_read_v1(uuid, text);
