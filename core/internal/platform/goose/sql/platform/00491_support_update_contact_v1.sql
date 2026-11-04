-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.support_update_contact_v1(
    arg_1 uuid,
    arg_2 bigint,
    arg_3 bigint,
    arg_4 uuid,
    arg_5 boolean,
    arg_6 text,
    arg_7 boolean,
    arg_8 text,
    arg_9 bigint,
    arg_10 text,
    arg_11 bytea,
    arg_12 timestamptz
)
RETURNS TABLE (
    response_document bytea,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        NULL::bytea,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.support_update_contact_v1(uuid, bigint, bigint, uuid, boolean, text, boolean, text, bigint, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_update_contact_v1(uuid, bigint, bigint, uuid, boolean, text, boolean, text, bigint, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_update_contact_v1(uuid, bigint, bigint, uuid, boolean, text, boolean, text, bigint, text, bytea, timestamptz);
