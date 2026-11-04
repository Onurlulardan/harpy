-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_release_hold_v2(
    arg_1 uuid,
    arg_2 bytea
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

REVOKE ALL ON FUNCTION documents_recovery_release_hold_v2(uuid, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_release_hold_v2(uuid, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_release_hold_v2(uuid, bytea);
