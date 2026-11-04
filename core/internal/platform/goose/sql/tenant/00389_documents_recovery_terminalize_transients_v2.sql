-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_terminalize_transients_v2(
    arg_1 uuid
)
RETURNS TABLE (
    cancelled_uploads bigint,
    expired_archives bigint,
    cancelled_archives bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_recovery_terminalize_transients_v2(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_terminalize_transients_v2(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_recovery_terminalize_transients_v2(uuid);
