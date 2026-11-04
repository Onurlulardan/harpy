-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_park_stale_outbox_v1(
    in_event_id uuid,
    in_operation_id uuid
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_park_stale_outbox_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_park_stale_outbox_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_park_stale_outbox_v1(uuid, uuid);
