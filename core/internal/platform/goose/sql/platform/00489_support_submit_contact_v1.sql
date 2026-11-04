-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.support_submit_contact_v1(
    arg_1 uuid,
    arg_2 text,
    arg_3 text,
    arg_4 text,
    arg_5 text,
    arg_6 text,
    arg_7 text,
    arg_8 text,
    arg_9 text,
    arg_10 bytea,
    arg_11 text,
    arg_12 bytea,
    arg_13 timestamptz
)
RETURNS TABLE (
    response_document bytea,
    replayed boolean,
    admitted boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        NULL::bytea,
        false,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.support_submit_contact_v1(uuid, text, text, text, text, text, text, text, text, bytea, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_submit_contact_v1(uuid, text, text, text, text, text, text, text, text, bytea, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_submit_contact_v1(uuid, text, text, text, text, text, text, text, text, bytea, text, bytea, timestamptz);
