-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_project_identity_record_v1(
    p_target_table text,
    p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    CASE p_target_table
    WHEN 'iam_profile_schema_versions' THEN
        INSERT INTO iam_profile_schema_versions
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_profile_schema_versions, p_record)).*;
    WHEN 'iam_profile_fields' THEN
        INSERT INTO iam_profile_fields
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_profile_fields, p_record)).*;
    WHEN 'iam_profile_field_revisions' THEN
        INSERT INTO iam_profile_field_revisions
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_profile_field_revisions, p_record)).*;
    WHEN 'iam_organizations' THEN
        INSERT INTO iam_organizations
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_organizations, p_record)).*;
    WHEN 'iam_users' THEN
        INSERT INTO iam_users
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_users, p_record)).*;
    WHEN 'iam_user_identifiers' THEN
        INSERT INTO iam_user_identifiers
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_user_identifiers, p_record)).*;
    WHEN 'iam_password_credentials' THEN
        INSERT INTO iam_password_credentials
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_password_credentials, p_record)).*;
    WHEN 'iam_organization_memberships' THEN
        INSERT INTO iam_organization_memberships
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_organization_memberships, p_record)).*;
    WHEN 'iam_roles' THEN
        INSERT INTO iam_roles
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_roles, p_record)).*;
    WHEN 'iam_user_role_assignments' THEN
        INSERT INTO iam_user_role_assignments
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_user_role_assignments, p_record)).*;
    WHEN 'iam_user_access_modes' THEN
        INSERT INTO iam_user_access_modes
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_user_access_modes, p_record)).*;
    WHEN 'iam_portal_subject_bindings' THEN
        INSERT INTO iam_portal_subject_bindings
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_portal_subject_bindings, p_record)).*;
    WHEN 'iam_portal_access_setups' THEN
        INSERT INTO iam_portal_access_setups
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_portal_access_setups, p_record)).*;
    ELSE
        RETURN false;
    END CASE;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_project_identity_record_v1(
    text, jsonb
) FROM PUBLIC;

