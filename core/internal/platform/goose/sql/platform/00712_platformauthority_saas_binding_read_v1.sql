-- +goose Up
CREATE FUNCTION platform.platformauthority_saas_binding_read_v1(
    in_environment_id uuid
)
RETURNS TABLE (
    binding_id text,
    environment_id text,
    mode text,
    kind text,
    license_id text,
    finalization text,
    recovery text,
    signer_generation bigint,
    signer_kid text,
    signer_public_key_sha256 text,
    signer_status text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        NULL::text,
        NULL::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_saas_binding_read_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_saas_binding_read_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_saas_binding_read_v1(uuid);
