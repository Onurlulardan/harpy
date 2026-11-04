-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_delivery_delete_artifact_v1(
    in_artifact uuid,
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::void;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_delivery_delete_artifact_v1(uuid, text, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_lifecycle_delivery_delete_artifact_v1(uuid, text, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_lifecycle_delivery_delete_artifact_v1(uuid, text, uuid, text);
