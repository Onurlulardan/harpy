-- +goose Up
DROP FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name);

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_assert_v2(in_operation_id uuid, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE ready boolean := false;
BEGIN
    IF in_schema_name::text !~ '^tenant_[0-9a-f]{32}$' THEN RETURN false; END IF;
    IF NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, in_schema_name) THEN RETURN false; END IF;
    EXECUTE pg_catalog.format('SELECT true FROM %I.system_resource_layout_installation_baselines WHERE operation_id = $1 AND tenant_id = $2', in_schema_name)
        INTO ready USING in_operation_id, in_tenant_id;
    RETURN COALESCE(ready, false)
        OR platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'systemresources.initialize_tenant_baseline', in_tenant_id, in_schema_name);
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name);
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_assert_v2(in_operation_id uuid, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) TO harpy_api, harpy_worker;
