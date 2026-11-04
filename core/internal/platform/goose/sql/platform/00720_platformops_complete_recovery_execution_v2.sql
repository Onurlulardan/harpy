-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_complete_recovery_execution_v2(
    in_recovery_set_id text,
    in_attempt_id text,
    in_lease_token text,
    in_manifest_sha256 text,
    in_manifest_bytes bytea,
    in_database_system_id text
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

REVOKE ALL ON FUNCTION platform.platformops_complete_recovery_execution_v2(text, text, text, text, bytea, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_complete_recovery_execution_v2(text, text, text, text, bytea, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_complete_recovery_execution_v2(text, text, text, text, bytea, text);
