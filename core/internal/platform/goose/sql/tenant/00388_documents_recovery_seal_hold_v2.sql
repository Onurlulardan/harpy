-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_seal_hold_v2(
    arg_1 uuid,
    arg_2 bytea,
    arg_3 bigint
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

REVOKE ALL ON FUNCTION documents_recovery_seal_hold_v2(uuid, bytea, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_seal_hold_v2(uuid, bytea, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_seal_hold_v2(uuid, bytea, bigint);
