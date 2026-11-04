-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_observe_v1(
    in_environment_id uuid,
    in_restore_operation_id uuid
)
RETURNS TABLE (
    state text,
    receipt_sha256 text,
    reactivated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        NULL::text,
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_observe_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_observe_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_observe_v1(uuid, uuid);
