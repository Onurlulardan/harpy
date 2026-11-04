-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_reactivation_fail_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_restore_operation_id uuid,
    in_occurred_at timestamptz
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

REVOKE ALL ON FUNCTION platform.iam_post_restore_reactivation_fail_v1(uuid, uuid, uuid, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_reactivation_fail_v1(uuid, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_reactivation_fail_v1(uuid, uuid, uuid, timestamptz);
