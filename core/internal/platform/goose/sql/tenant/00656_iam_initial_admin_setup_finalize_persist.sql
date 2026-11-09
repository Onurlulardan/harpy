-- +goose Up
DROP FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text);

-- +goose StatementBegin
CREATE FUNCTION iam_initial_admin_setup_finalize_v2(in_schema_name name, in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_manifest_sha256 text)
RETURNS uuid LANGUAGE plpgsql STABLE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
DECLARE tenant_schema name; identity_tenant uuid; identity_schema text; artifact uuid;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    IF tenant_schema IS DISTINCT FROM in_schema_name THEN RETURN '00000000-0000-0000-0000-000000000000'::uuid; END IF;
    EXECUTE pg_catalog.format('SELECT tenant_id, schema_name FROM %I.tenancy_schema_identity WHERE singleton', tenant_schema) INTO identity_tenant, identity_schema;
    IF identity_tenant IS DISTINCT FROM in_tenant_id OR identity_schema IS DISTINCT FROM tenant_schema::text THEN
        RETURN '00000000-0000-0000-0000-000000000000'::uuid;
    END IF;
    SELECT artifact_id INTO artifact FROM platform.iam_initial_admin_setup_stages
    WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    RETURN COALESCE(artifact, '00000000-0000-0000-0000-000000000000'::uuid);
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text);
-- +goose StatementBegin
CREATE FUNCTION iam_initial_admin_setup_finalize_v2(in_schema_name name, in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_manifest_sha256 text)
RETURNS uuid LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT '00000000-0000-0000-0000-000000000000'::uuid; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v2(name, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;
