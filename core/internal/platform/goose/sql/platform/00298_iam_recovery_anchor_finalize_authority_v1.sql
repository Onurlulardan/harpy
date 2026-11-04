-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(
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

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(uuid);
