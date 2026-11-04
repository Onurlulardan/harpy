-- +goose Up
CREATE FUNCTION iam_invitation_delivery_worker_release_v1(
    in_schema_name name,
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text
)
RETURNS TABLE (
    executed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_release_v1(name, text, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_release_v1(name, text, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_delivery_worker_release_v1(name, text, uuid, text);
