-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_query_publish_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 bytea,
    arg_6 uuid,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 bigint,
    arg_10 bigint,
    arg_11 bytea,
    arg_12 bytea,
    arg_13 bytea,
    arg_14 jsonb,
    arg_15 bytea,
    arg_16 text,
    arg_17 timestamptz
)
RETURNS TABLE (
    query_id uuid,
    version_id uuid,
    version_no bigint,
    root_revision bigint,
    reference_epoch bigint,
    version_revision bigint,
    lifecycle_state text,
    version_state text,
    "name" text,
    query_document bytea,
    parameter_document bytea,
    output_document bytea,
    manifest_sha256 bytea,
    cache_policy text,
    created_at timestamptz,
    updated_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        ''::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_publish_v1(name, uuid, uuid, text, bytea, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, jsonb, bytea, text, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_publish_v1(name, uuid, uuid, text, bytea, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, jsonb, bytea, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_query_publish_v1(name, uuid, uuid, text, bytea, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, jsonb, bytea, text, timestamptz);
