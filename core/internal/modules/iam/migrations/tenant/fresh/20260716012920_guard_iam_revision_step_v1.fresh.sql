-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_revision_step_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_NARGS <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'IAM revision guard invocation is invalid';
    END IF;
    IF NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = TG_TABLE_NAME || '_revision_guard',
            MESSAGE = 'IAM revision must increase by exactly one';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_revision_step_v1() FROM PUBLIC;

CREATE TRIGGER iam_users_revision_guard
BEFORE UPDATE ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_revision_guard;

CREATE TRIGGER iam_user_identifiers_revision_guard
BEFORE UPDATE ON iam_user_identifiers
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_user_identifiers
    ENABLE ALWAYS TRIGGER iam_user_identifiers_revision_guard;

CREATE TRIGGER iam_organizations_revision_guard
BEFORE UPDATE ON iam_organizations
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_organizations
    ENABLE ALWAYS TRIGGER iam_organizations_revision_guard;

CREATE TRIGGER iam_password_credentials_revision_guard
BEFORE UPDATE ON iam_password_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_password_credentials
    ENABLE ALWAYS TRIGGER iam_password_credentials_revision_guard;

CREATE TRIGGER iam_organization_memberships_revision_guard
BEFORE UPDATE ON iam_organization_memberships
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_organization_memberships
    ENABLE ALWAYS TRIGGER iam_organization_memberships_revision_guard;

CREATE TRIGGER iam_roles_revision_guard
BEFORE UPDATE ON iam_roles
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_revision_guard;

CREATE TRIGGER iam_user_role_assignments_revision_guard
BEFORE UPDATE ON iam_user_role_assignments
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_user_role_assignments
    ENABLE ALWAYS TRIGGER iam_user_role_assignments_revision_guard;

CREATE TRIGGER iam_user_access_modes_revision_guard
BEFORE UPDATE ON iam_user_access_modes
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_user_access_modes
    ENABLE ALWAYS TRIGGER iam_user_access_modes_revision_guard;

CREATE TRIGGER iam_portal_subject_bindings_revision_guard
BEFORE UPDATE ON iam_portal_subject_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_portal_subject_bindings
    ENABLE ALWAYS TRIGGER iam_portal_subject_bindings_revision_guard;

CREATE TRIGGER iam_portal_access_setups_revision_guard
BEFORE UPDATE ON iam_portal_access_setups
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_portal_access_setups
    ENABLE ALWAYS TRIGGER iam_portal_access_setups_revision_guard;

CREATE TRIGGER iam_sessions_revision_guard
BEFORE UPDATE ON iam_sessions
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_sessions ENABLE ALWAYS TRIGGER iam_sessions_revision_guard;

CREATE TRIGGER iam_authenticators_revision_guard
BEFORE UPDATE ON iam_authenticators
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_authenticators
    ENABLE ALWAYS TRIGGER iam_authenticators_revision_guard;

CREATE TRIGGER iam_webauthn_credentials_revision_guard
BEFORE UPDATE ON iam_webauthn_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_webauthn_credentials
    ENABLE ALWAYS TRIGGER iam_webauthn_credentials_revision_guard;

CREATE TRIGGER iam_totp_authenticators_revision_guard
BEFORE UPDATE ON iam_totp_authenticators
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_totp_authenticators
    ENABLE ALWAYS TRIGGER iam_totp_authenticators_revision_guard;

CREATE TRIGGER iam_recovery_codes_revision_guard
BEFORE UPDATE ON iam_recovery_codes
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_recovery_codes
    ENABLE ALWAYS TRIGGER iam_recovery_codes_revision_guard;

CREATE TRIGGER iam_auth_challenges_revision_guard
BEFORE UPDATE ON iam_auth_challenges
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_auth_challenges
    ENABLE ALWAYS TRIGGER iam_auth_challenges_revision_guard;

CREATE TRIGGER iam_invitations_revision_guard
BEFORE UPDATE ON iam_invitations
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_invitations
    ENABLE ALWAYS TRIGGER iam_invitations_revision_guard;

CREATE TRIGGER iam_one_time_tokens_revision_guard
BEFORE UPDATE ON iam_one_time_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_one_time_tokens
    ENABLE ALWAYS TRIGGER iam_one_time_tokens_revision_guard;

CREATE TRIGGER iam_refresh_families_revision_guard
BEFORE UPDATE ON iam_refresh_families
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_refresh_families
    ENABLE ALWAYS TRIGGER iam_refresh_families_revision_guard;

CREATE TRIGGER iam_refresh_tokens_revision_guard
BEFORE UPDATE ON iam_refresh_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_refresh_tokens
    ENABLE ALWAYS TRIGGER iam_refresh_tokens_revision_guard;
