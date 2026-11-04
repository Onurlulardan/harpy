-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_record_copy_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 text,
    arg_4 text,
    arg_5 text,
    arg_6 text,
    arg_7 bytea,
    arg_8 bigint,
    arg_9 text
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_recovery_record_copy_v1(uuid, uuid, text, text, text, text, bytea, bigint, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_record_copy_v1(uuid, uuid, text, text, text, text, bytea, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_record_copy_v1(uuid, uuid, text, text, text, text, bytea, bigint, text);
