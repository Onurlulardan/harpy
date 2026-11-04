-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_notifications_session_selector_v1(
    in_schema_name name,
    in_principal_id uuid,
    in_session_id uuid,
    in_utc timestamptz
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT ''::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_notifications_session_selector_v1(name, uuid, uuid, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_notifications_session_selector_v1(name, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_notifications_session_selector_v1(name, uuid, uuid, timestamptz);
