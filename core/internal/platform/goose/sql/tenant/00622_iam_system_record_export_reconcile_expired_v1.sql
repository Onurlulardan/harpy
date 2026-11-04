-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_expired_v1(
    in_schema_name name,
    in_tenant_id uuid
)
RETURNS TABLE (
    recovered boolean,
    resolution text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        false,
        NULL::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_expired_v1(name, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_expired_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_reconcile_expired_v1(name, uuid);
