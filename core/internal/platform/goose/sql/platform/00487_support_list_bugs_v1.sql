-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.support_list_bugs_v1(
    arg_1 uuid,
    arg_2 bigint,
    arg_3 bigint,
    arg_4 boolean,
    arg_5 uuid,
    arg_6 uuid,
    arg_7 text,
    arg_8 text,
    arg_9 text,
    arg_10 integer,
    arg_11 integer
)
RETURNS TABLE (
    items bytea,
    total bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        NULL::bytea,
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.support_list_bugs_v1(uuid, bigint, bigint, boolean, uuid, uuid, text, text, text, integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_list_bugs_v1(uuid, bigint, bigint, boolean, uuid, uuid, text, text, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_list_bugs_v1(uuid, bigint, bigint, boolean, uuid, uuid, text, text, text, integer, integer);
