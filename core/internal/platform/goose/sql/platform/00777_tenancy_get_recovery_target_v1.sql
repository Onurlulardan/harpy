-- +goose Up
-- BELIRSIZ: no recovery target is stored. tenant_set_sha256 is 32 bytes so Scan can copy it.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_recovery_target_v1(
    p_purpose text
)
RETURNS TABLE (
    environment_id uuid,
    tenant_id uuid,
    tenant_slug text,
    schema_name text,
    tenant_set_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        decode('0000000000000000000000000000000000000000000000000000000000000000', 'hex');
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_recovery_target_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_recovery_target_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_get_recovery_target_v1(text);
