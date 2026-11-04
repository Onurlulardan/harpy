-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_reconcile_event_gap_v2(
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint
)
RETURNS TABLE (
    completed_at timestamptz,
    requeued boolean,
    action_id uuid,
    disposition text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        pg_catalog.clock_timestamp(),
        false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_reconcile_event_gap_v2(uuid, uuid, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_reconcile_event_gap_v2(uuid, uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_reconcile_event_gap_v2(uuid, uuid, uuid, bigint);
