-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_start_recovery_snapshot_v1(
    in_recovery_set_id text,
    in_attempt_id text,
    in_lease_token text
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

REVOKE ALL ON FUNCTION platform.platformops_start_recovery_snapshot_v1(text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_start_recovery_snapshot_v1(text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_start_recovery_snapshot_v1(text, text, text);
