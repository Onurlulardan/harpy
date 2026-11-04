-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_event_sequence_v1(
    in_operation_id uuid,
    in_command_key text,
    in_aggregate_kind text,
    in_aggregate_id uuid
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_event_sequence_v1(uuid, text, text, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_lifecycle_event_sequence_v1(uuid, text, text, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_lifecycle_event_sequence_v1(uuid, text, text, uuid);
