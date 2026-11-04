-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(
    in_schema_name name,
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    executed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(name, text, uuid, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(name, text, uuid, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(name, text, uuid, text, timestamptz);
