-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_execute_retention_cycle_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 text,
    arg_7 text,
    arg_8 integer
)
RETURNS TABLE (
    job_id uuid,
    run_at timestamptz,
    payload_reference text,
    payload_sha256 text,
    dedupe_key text,
    priority integer,
    drafts bigint,
    versions bigint,
    releases bigint,
    executions bigint,
    audits bigint,
    capabilities bigint,
    legal_hold_blocked boolean,
    has_remaining boolean,
    checkpoint_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        pg_catalog.clock_timestamp(),
        ''::text,
        ''::text,
        ''::text,
        0,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        false,
        false,
        0::bigint;
$$;

REVOKE ALL ON FUNCTION analytics_execute_retention_cycle_v1(name, uuid, uuid, uuid, uuid, text, text, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_execute_retention_cycle_v1(name, uuid, uuid, uuid, uuid, text, text, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_execute_retention_cycle_v1(name, uuid, uuid, uuid, uuid, text, text, integer);
