-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_immutable_columns_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    column_index integer;
    column_name text;
    old_row jsonb;
    new_row jsonb;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_NARGS < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'IAM immutable guard invocation is invalid';
    END IF;
    old_row := pg_catalog.to_jsonb(OLD);
    new_row := pg_catalog.to_jsonb(NEW);
    FOR column_index IN 0..TG_NARGS - 1 LOOP
        column_name := TG_ARGV[column_index];
        IF (new_row -> column_name) IS DISTINCT FROM (old_row -> column_name) THEN
            RAISE EXCEPTION USING
                ERRCODE = '23514',
                CONSTRAINT = TG_TABLE_NAME || '_immutable_guard',
                MESSAGE = pg_catalog.format(
                    'IAM immutable column %I cannot change',
                    column_name
                );
        END IF;
    END LOOP;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_immutable_columns_v1() FROM PUBLIC;

CREATE TRIGGER iam_users_immutable_guard
BEFORE UPDATE ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'identity_kind', 'profile_resource_kind', 'created_at'
);
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_immutable_guard;

CREATE TRIGGER iam_user_identifiers_immutable_guard
BEFORE UPDATE ON iam_user_identifiers
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'kind', 'normalized_value', 'normalized_sha256',
    'created_at'
);
ALTER TABLE iam_user_identifiers
    ENABLE ALWAYS TRIGGER iam_user_identifiers_immutable_guard;

CREATE TRIGGER iam_organizations_immutable_guard
BEFORE UPDATE ON iam_organizations
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'created_at'
);
ALTER TABLE iam_organizations
    ENABLE ALWAYS TRIGGER iam_organizations_immutable_guard;

CREATE TRIGGER iam_password_credentials_immutable_guard
BEFORE UPDATE ON iam_password_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'user_id', 'created_at'
);
ALTER TABLE iam_password_credentials
    ENABLE ALWAYS TRIGGER iam_password_credentials_immutable_guard;

CREATE TRIGGER iam_organization_memberships_immutable_guard
BEFORE UPDATE ON iam_organization_memberships
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'organization_id', 'user_id', 'created_at'
);
ALTER TABLE iam_organization_memberships
    ENABLE ALWAYS TRIGGER iam_organization_memberships_immutable_guard;

CREATE TRIGGER iam_roles_immutable_guard
BEFORE UPDATE ON iam_roles
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'scope_kind', 'scope_id', 'organization_id', 'created_at'
);
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_immutable_guard;

CREATE TRIGGER iam_user_role_assignments_immutable_guard
BEFORE UPDATE ON iam_user_role_assignments
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'role_id', 'role_scope_kind', 'role_scope_id',
    'organization_id', 'membership_id', 'created_at'
);
ALTER TABLE iam_user_role_assignments
    ENABLE ALWAYS TRIGGER iam_user_role_assignments_immutable_guard;

CREATE TRIGGER iam_user_access_modes_immutable_guard
BEFORE UPDATE ON iam_user_access_modes
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'mode', 'created_at'
);
ALTER TABLE iam_user_access_modes
    ENABLE ALWAYS TRIGGER iam_user_access_modes_immutable_guard;

CREATE TRIGGER iam_portal_subject_bindings_immutable_guard
BEFORE UPDATE ON iam_portal_subject_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'access_mode_id', 'access_mode', 'subject_shape',
    'subject_type', 'profile_owner', 'profile_form_id',
    'profile_form_version_id', 'subject_record_id', 'identifier_field_id',
    'contract_sha256', 'created_at'
);
ALTER TABLE iam_portal_subject_bindings
    ENABLE ALWAYS TRIGGER iam_portal_subject_bindings_immutable_guard;

CREATE TRIGGER iam_portal_access_setups_immutable_guard
BEFORE UPDATE ON iam_portal_access_setups
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'binding_id', 'user_id', 'access_mode_id', 'access_mode', 'provenance',
    'source_evidence_sha256', 'created_at'
);
ALTER TABLE iam_portal_access_setups
    ENABLE ALWAYS TRIGGER iam_portal_access_setups_immutable_guard;
