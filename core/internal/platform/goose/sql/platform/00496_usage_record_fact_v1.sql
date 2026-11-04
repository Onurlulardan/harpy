-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.usage_record_fact_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 bytea,
    arg_6 bytea,
    arg_7 text,
    arg_8 text,
    arg_9 integer,
    arg_10 bigint,
    arg_11 timestamptz
)
RETURNS TABLE (
    fact_id uuid,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.usage_record_fact_v1(uuid, uuid, uuid, uuid, bytea, bytea, text, text, integer, bigint, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.usage_record_fact_v1(uuid, uuid, uuid, uuid, bytea, bytea, text, text, integer, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.usage_record_fact_v1(uuid, uuid, uuid, uuid, bytea, bytea, text, text, integer, bigint, timestamptz);
