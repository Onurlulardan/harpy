-- +goose Up
CREATE TABLE IF NOT EXISTS system_resource_layout_installation_baselines (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_revision_id uuid NOT NULL,
    organization_revision_id uuid NOT NULL,
    role_revision_id uuid NOT NULL,
    baseline_sha256 bytea NOT NULL,
    initialized_at timestamptz NOT NULL,
    CONSTRAINT system_resource_layout_installation_baselines_sha_length CHECK (octet_length(baseline_sha256) = 32)
);

REVOKE ALL ON TABLE system_resource_layout_installation_baselines FROM PUBLIC, harpy_api, harpy_worker;
-- +goose StatementBegin
DO $grant$ BEGIN
    EXECUTE pg_catalog.format('GRANT USAGE ON SCHEMA %I TO harpy_owner', pg_catalog.current_schema());
    EXECUTE pg_catalog.format('GRANT SELECT ON TABLE %I.system_resource_layout_installation_baselines TO harpy_owner', pg_catalog.current_schema());
END $grant$;
-- +goose StatementEnd

DROP FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name);

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_record_v2(
    in_operation_id uuid, in_tenant_id uuid, in_user_revision_id uuid, in_organization_revision_id uuid, in_role_revision_id uuid,
    in_initialized_at timestamptz, in_schema_name name
) RETURNS bytea LANGUAGE plpgsql VOLATILE SECURITY DEFINER
AS $fn$
DECLARE tenant_schema name; digest bytea; published integer;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    IF tenant_schema IS DISTINCT FROM in_schema_name OR in_initialized_at IS NULL THEN
        RAISE EXCEPTION 'system resource baseline is not bound to this tenant' USING ERRCODE = '42501';
    END IF;
    IF NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, tenant_schema) THEN
        RAISE EXCEPTION 'system resource baseline is outside its activation' USING ERRCODE = '42501';
    END IF;
    digest := pg_catalog.sha256(pg_catalog.convert_to('systemresources.installation_baseline|' || in_operation_id::text || '|' || in_tenant_id::text || '|' || in_user_revision_id::text || '|' || in_organization_revision_id::text || '|' || in_role_revision_id::text, 'UTF8'));
    EXECUTE pg_catalog.format('SELECT baseline_sha256 FROM %I.system_resource_layout_installation_baselines WHERE operation_id = $1 AND tenant_id = $2 AND user_revision_id = $3 AND organization_revision_id = $4 AND role_revision_id = $5', tenant_schema)
        INTO digest USING in_operation_id, in_tenant_id, in_user_revision_id, in_organization_revision_id, in_role_revision_id;
    IF digest IS NOT NULL AND pg_catalog.octet_length(digest) = 32 THEN RETURN digest; END IF;
    EXECUTE pg_catalog.format('SELECT count(*) FROM %I.system_resource_layout_revisions WHERE id IN ($1,$2,$3) AND state = ''published'' AND row_revision > 1', tenant_schema)
        INTO published USING in_user_revision_id, in_organization_revision_id, in_role_revision_id;
    IF published <> 3 THEN RAISE EXCEPTION 'system resource baseline revisions are not published' USING ERRCODE = '23514'; END IF;
    digest := pg_catalog.sha256(pg_catalog.convert_to('systemresources.installation_baseline|' || in_operation_id::text || '|' || in_tenant_id::text || '|' || in_user_revision_id::text || '|' || in_organization_revision_id::text || '|' || in_role_revision_id::text, 'UTF8'));
    EXECUTE pg_catalog.format('INSERT INTO %I.system_resource_layout_installation_baselines (operation_id, tenant_id, user_revision_id, organization_revision_id, role_revision_id, baseline_sha256, initialized_at) VALUES ($1,$2,$3,$4,$5,$6,$7)', tenant_schema)
        USING in_operation_id, in_tenant_id, in_user_revision_id, in_organization_revision_id, in_role_revision_id, digest, in_initialized_at;
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'systemresources.initialize_tenant_baseline', in_tenant_id, tenant_schema);
    RETURN digest;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name);
DROP TABLE IF EXISTS system_resource_layout_installation_baselines;
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_record_v2(in_operation_id uuid, in_tenant_id uuid, in_user_revision_id uuid, in_organization_revision_id uuid, in_role_revision_id uuid, in_initialized_at timestamptz, in_schema_name name)
RETURNS bytea LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$ SELECT '\x0000000000000000000000000000000000000000000000000000000000000000'::bytea; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_record_v2(uuid, uuid, uuid, uuid, uuid, timestamptz, name) TO harpy_api, harpy_worker;
