-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.support_create_bug_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 text,
    arg_4 text,
    arg_5 uuid,
    arg_6 text,
    arg_7 text,
    arg_8 bigint,
    arg_9 bigint,
    arg_10 text,
    arg_11 text,
    arg_12 text,
    arg_13 text,
    arg_14 text,
    arg_15 text,
    arg_16 jsonb,
    arg_17 jsonb,
    arg_18 jsonb,
    arg_19 text,
    arg_20 text,
    arg_21 bytea,
    arg_22 timestamptz
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

REVOKE ALL ON FUNCTION platform.support_create_bug_v1(uuid, uuid, text, text, uuid, text, text, bigint, bigint, text, text, text, text, text, text, jsonb, jsonb, jsonb, text, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_create_bug_v1(uuid, uuid, text, text, uuid, text, text, bigint, bigint, text, text, text, text, text, text, jsonb, jsonb, jsonb, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_create_bug_v1(uuid, uuid, text, text, uuid, text, text, bigint, bigint, text, text, text, text, text, text, jsonb, jsonb, jsonb, text, text, bytea, timestamptz);
