-- +goose Up
CREATE FUNCTION platform.platform_async_dispose_stale_job_v1(
    in_job_id uuid,
    in_decision text,
    in_root_id uuid,
    in_action_id uuid,
    in_result_job_id uuid,
    in_result_root_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_group_id text,
    in_group_generation bigint,
    in_domain text,
    in_generation bigint,
    in_result_payload_reference text,
    in_result_dedupe_key text,
    in_canonical bytea,
    in_digest text,
    in_reason text,
    in_ticket text,
    in_actor_id uuid,
    in_causation_id uuid,
    in_source_run_at text,
    in_result_run_at text,
    in_occurred_at text
)
RETURNS TABLE (
    completed_at timestamptz,
    action_id uuid,
    result_job_id uuid,
    result_root_id uuid,
    created boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        pg_catalog.clock_timestamp(),
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        NULL::uuid,
        false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_dispose_stale_job_v1(uuid, text, uuid, uuid, uuid, uuid, text, uuid, text, bigint, text, bigint, text, text, bytea, text, text, text, uuid, uuid, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_dispose_stale_job_v1(uuid, text, uuid, uuid, uuid, uuid, text, uuid, text, bigint, text, bigint, text, text, bytea, text, text, text, uuid, uuid, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_dispose_stale_job_v1(uuid, text, uuid, uuid, uuid, uuid, text, uuid, text, bigint, text, bigint, text, text, bytea, text, text, text, uuid, uuid, text, text, text);
