-- +goose Up
CREATE TABLE IF NOT EXISTS iam_installation_identity_bindings (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    protected_assignment_id uuid NOT NULL,
    identifier_id uuid NOT NULL,
    membership_id uuid NOT NULL,
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    role_permission_id uuid NOT NULL,
    binding_set_sha256 text NOT NULL
);

REVOKE ALL ON TABLE iam_installation_identity_bindings FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid);

-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v4(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_user_id uuid, in_organization_id uuid,
    in_protected_assignment_id uuid, in_email text, in_display_name text, in_tenant_name text, in_binding_set_sha256 text,
    in_user_schema_id uuid, in_user_field_id uuid, in_organization_schema_id uuid, in_organization_field_id uuid,
    in_role_schema_id uuid, in_role_field_id uuid, in_identifier_id uuid, in_access_mode_id uuid, in_membership_id uuid,
    in_role_id uuid, in_permission_id uuid, in_role_permission_id uuid
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER
AS $fn$
DECLARE tenant_schema name; identity_tenant uuid; already boolean; resource_id uuid; action_id uuid;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    IF tenant_schema::text !~ '^tenant_[0-9a-f]{32}$' THEN
        RAISE EXCEPTION 'tenant schema is not bound' USING ERRCODE = '42501';
    END IF;
    EXECUTE pg_catalog.format('SELECT tenant_id FROM %I.tenancy_schema_identity WHERE singleton', tenant_schema) INTO identity_tenant;
    IF identity_tenant IS DISTINCT FROM in_tenant_id THEN
        RAISE EXCEPTION 'tenant schema identity does not match the installation' USING ERRCODE = '42501';
    END IF;
    IF in_email IS NULL OR in_display_name IS NULL OR in_tenant_name IS NULL OR in_binding_set_sha256 IS NULL
        OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, tenant_schema) THEN
        RAISE EXCEPTION 'IAM installation identity is outside its activation' USING ERRCODE = '42501';
    END IF;
    EXECUTE pg_catalog.format('SELECT true FROM %I.iam_installation_identity_bindings WHERE operation_id = $1 AND user_id = $2 AND organization_id = $3', tenant_schema) INTO already USING in_operation_id, in_user_id, in_organization_id;
    IF already THEN RETURN true; END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'iam.initialize_tenant_identity', in_tenant_id, tenant_schema) THEN
        RAISE EXCEPTION 'IAM installation identity capability is missing' USING ERRCODE = '42501';
    END IF;
    resource_id := in_role_schema_id;
    action_id := in_role_field_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_users (id, revision) VALUES ($1, 1) ON CONFLICT (id) DO NOTHING', tenant_schema) USING in_user_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_organizations (id, status, revision, slug) VALUES ($1, ''active'', 1, ''root'') ON CONFLICT (id) DO NOTHING', tenant_schema) USING in_organization_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_resources (id, resource_key, state) VALUES ($1, ''installation.tenant'', ''active'') ON CONFLICT (id) DO NOTHING', tenant_schema) USING resource_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_actions (id, action_key, state) VALUES ($1, ''installation.administer'', ''active'') ON CONFLICT (id) DO NOTHING', tenant_schema) USING action_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_permissions (id, resource_id, action_id, match_kind, is_protected, state) VALUES ($1, $2, $3, ''exact'', true, ''active'') ON CONFLICT (id) DO NOTHING', tenant_schema) USING in_permission_id, resource_id, action_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_roles (id, revision, scope_kind, scope_id, name_key, status, is_protected) VALUES ($1, 1, ''organization'', $2, ''protected-admin'', ''active'', true) ON CONFLICT (id) DO NOTHING', tenant_schema) USING in_role_id, in_organization_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_user_identifiers (id, revision, state, cooldown_until, normalized_value, user_id, kind) VALUES ($1, 1, ''active'', '''', $2, $3, ''email'') ON CONFLICT (id) DO NOTHING', tenant_schema) USING in_identifier_id, in_email, in_user_id;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_installation_identity_bindings (operation_id, tenant_id, user_id, organization_id, protected_assignment_id, identifier_id, membership_id, role_id, permission_id, role_permission_id, binding_set_sha256) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)', tenant_schema)
        USING in_operation_id, in_tenant_id, in_user_id, in_organization_id, in_protected_assignment_id, in_identifier_id, in_membership_id, in_role_id, in_permission_id, in_role_permission_id, in_binding_set_sha256;
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid);
DROP TABLE IF EXISTS iam_installation_identity_bindings;
-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v4(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_user_id uuid, in_organization_id uuid, in_protected_assignment_id uuid, in_email text, in_display_name text, in_tenant_name text, in_binding_set_sha256 text, in_user_schema_id uuid, in_user_field_id uuid, in_organization_schema_id uuid, in_organization_field_id uuid, in_role_schema_id uuid, in_role_field_id uuid, in_identifier_id uuid, in_access_mode_id uuid, in_membership_id uuid, in_role_id uuid, in_permission_id uuid, in_role_permission_id uuid)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v4(uuid, uuid, uuid, uuid, uuid, uuid, text, text, text, text, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid) TO harpy_api, harpy_worker;
