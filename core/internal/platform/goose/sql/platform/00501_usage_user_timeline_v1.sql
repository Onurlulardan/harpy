-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION platform.usage_user_timeline_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 bigint,
    arg_6 bigint,
    arg_7 bigint,
    arg_8 text,
    arg_9 text,
    arg_10 text,
    arg_11 boolean,
    arg_12 text,
    arg_13 text,
    arg_14 text,
    arg_15 integer,
    arg_16 integer
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::bytea;
$$;

REVOKE ALL ON FUNCTION platform.usage_user_timeline_v1(uuid, uuid, uuid, uuid, bigint, bigint, bigint, text, text, text, boolean, text, text, text, integer, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.usage_user_timeline_v1(uuid, uuid, uuid, uuid, bigint, bigint, bigint, text, text, text, boolean, text, text, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.usage_user_timeline_v1(uuid, uuid, uuid, uuid, bigint, bigint, bigint, text, text, text, boolean, text, text, text, integer, integer);
