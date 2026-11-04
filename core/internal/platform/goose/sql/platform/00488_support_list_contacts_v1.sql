-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION platform.support_list_contacts_v1(
    arg_1 uuid,
    arg_2 bigint,
    arg_3 bigint,
    arg_4 text,
    arg_5 text,
    arg_6 integer,
    arg_7 integer
)
RETURNS TABLE (
    items bytea,
    total bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::bytea,
        0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.support_list_contacts_v1(uuid, bigint, bigint, text, text, integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_list_contacts_v1(uuid, bigint, bigint, text, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_list_contacts_v1(uuid, bigint, bigint, text, text, integer, integer);
