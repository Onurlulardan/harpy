-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_security_epoch_initialize_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
    in_observed_at timestamptz
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

REVOKE ALL ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz);
