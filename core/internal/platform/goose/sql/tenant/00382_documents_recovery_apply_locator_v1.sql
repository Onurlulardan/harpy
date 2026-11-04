-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_apply_locator_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 text,
    arg_6 text,
    arg_7 text,
    arg_8 bigint,
    arg_9 bytea
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

REVOKE ALL ON FUNCTION documents_recovery_apply_locator_v1(uuid, uuid, uuid, text, text, text, text, bigint, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_apply_locator_v1(uuid, uuid, uuid, text, text, text, text, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_apply_locator_v1(uuid, uuid, uuid, text, text, text, text, bigint, bytea);
