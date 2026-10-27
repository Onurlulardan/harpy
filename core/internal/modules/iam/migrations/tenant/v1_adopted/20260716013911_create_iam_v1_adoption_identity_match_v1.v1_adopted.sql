-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_identity_record_matches_v1(
    p_target_table text,
    p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_match boolean;
BEGIN
    CASE p_target_table
    WHEN 'iam_profile_schema_versions' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_profile_schema_versions, p_record))
        INTO v_match FROM iam_profile_schema_versions AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_profile_fields' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_profile_fields, p_record))
        INTO v_match FROM iam_profile_fields AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_profile_field_revisions' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_profile_field_revisions, p_record))
        INTO v_match FROM iam_profile_field_revisions AS actual
        WHERE actual.schema_version_id = (p_record->>'schema_version_id')::uuid
          AND actual.field_id = (p_record->>'field_id')::uuid;
    WHEN 'iam_organizations' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_organizations, p_record))
        INTO v_match FROM iam_organizations AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_users' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_users, p_record))
        INTO v_match FROM iam_users AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_user_identifiers' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_user_identifiers, p_record))
        INTO v_match FROM iam_user_identifiers AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_password_credentials' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_password_credentials, p_record))
        INTO v_match FROM iam_password_credentials AS actual
        WHERE actual.user_id = (p_record->>'user_id')::uuid;
    WHEN 'iam_organization_memberships' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_organization_memberships, p_record))
        INTO v_match FROM iam_organization_memberships AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_roles' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_roles, p_record))
        INTO v_match FROM iam_roles AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_user_role_assignments' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_user_role_assignments, p_record))
        INTO v_match FROM iam_user_role_assignments AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_user_access_modes' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_user_access_modes, p_record))
        INTO v_match FROM iam_user_access_modes AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_portal_subject_bindings' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_portal_subject_bindings, p_record))
        INTO v_match FROM iam_portal_subject_bindings AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_portal_access_setups' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_portal_access_setups, p_record))
        INTO v_match FROM iam_portal_access_setups AS actual
        WHERE actual.binding_id = (p_record->>'binding_id')::uuid;
    ELSE
        RETURN false;
    END CASE;
    RETURN COALESCE(v_match, false);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_identity_record_matches_v1(
    text, jsonb
) FROM PUBLIC;

