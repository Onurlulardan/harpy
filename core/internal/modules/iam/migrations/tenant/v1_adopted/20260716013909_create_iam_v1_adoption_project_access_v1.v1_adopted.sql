-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_project_access_record_v1(
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
    WHEN 'iam_resources' THEN
        INSERT INTO iam_resources
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_resources, p_record)).*;
    WHEN 'iam_actions' THEN
        INSERT INTO iam_actions
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_actions, p_record)).*;
    WHEN 'iam_permissions' THEN
        INSERT INTO iam_permissions
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_permissions, p_record)).*;
    WHEN 'iam_user_permission_bindings' THEN
        INSERT INTO iam_user_permission_bindings
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_user_permission_bindings, p_record)).*;
    WHEN 'iam_role_permission_bindings' THEN
        INSERT INTO iam_role_permission_bindings
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_role_permission_bindings, p_record)).*;
    WHEN 'iam_organization_permission_bindings' THEN
        INSERT INTO iam_organization_permission_bindings
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_organization_permission_bindings, p_record)).*;
    WHEN 'iam_authorization_clock' THEN
        INSERT INTO iam_authorization_clock
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_authorization_clock, p_record)).*;
    WHEN 'iam_protected_admin_policies' THEN
        INSERT INTO iam_protected_admin_policies
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_protected_admin_policies, p_record)).*;
    WHEN 'iam_permission_request_groups' THEN
        INSERT INTO iam_permission_request_groups
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_permission_request_groups, p_record)).*;
    WHEN 'iam_permission_requests' THEN
        INSERT INTO iam_permission_requests
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_permission_requests, p_record)).*;
    WHEN 'iam_invitations' THEN
        INSERT INTO iam_invitations
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_invitations, p_record)).*;
    WHEN 'iam_one_time_tokens' THEN
        INSERT INTO iam_one_time_tokens
        SELECT (pg_catalog.jsonb_populate_record(NULL::iam_one_time_tokens, p_record)).*;
    ELSE
        RETURN false;
    END CASE;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_project_access_record_v1(
    text, jsonb
) FROM PUBLIC;
