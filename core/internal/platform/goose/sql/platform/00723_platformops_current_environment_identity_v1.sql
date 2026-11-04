-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_current_environment_identity_v1(
    in_expected_environment_id text
)
RETURNS TABLE (
    environment_id text,
    mode text,
    primary_binding_id text,
    tenant_set_sha256 text,
    owner_binding_id text,
    owner_binding_generation bigint,
    database_lineage_id text,
    database_system_id text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_current_environment_identity_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_current_environment_identity_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_current_environment_identity_v1(text);
