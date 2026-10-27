-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_access_record_matches_v1(
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
    WHEN 'iam_resources' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_resources, p_record))
        INTO v_match FROM iam_resources AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_actions' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_actions, p_record))
        INTO v_match FROM iam_actions AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_permissions' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_permissions, p_record))
        INTO v_match FROM iam_permissions AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_user_permission_bindings' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_user_permission_bindings, p_record))
        INTO v_match FROM iam_user_permission_bindings AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_role_permission_bindings' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_role_permission_bindings, p_record))
        INTO v_match FROM iam_role_permission_bindings AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_organization_permission_bindings' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_organization_permission_bindings, p_record))
        INTO v_match FROM iam_organization_permission_bindings AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_authorization_clock' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_authorization_clock, p_record))
        INTO v_match FROM iam_authorization_clock AS actual
        WHERE actual.singleton = (p_record->>'singleton')::boolean;
    WHEN 'iam_protected_admin_policies' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_protected_admin_policies, p_record))
        INTO v_match FROM iam_protected_admin_policies AS actual
        WHERE actual.singleton = (p_record->>'singleton')::boolean;
    WHEN 'iam_permission_request_groups' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_permission_request_groups, p_record))
        INTO v_match FROM iam_permission_request_groups AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_permission_requests' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_permission_requests, p_record))
        INTO v_match FROM iam_permission_requests AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_invitations' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_invitations, p_record))
        INTO v_match FROM iam_invitations AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    WHEN 'iam_one_time_tokens' THEN
        SELECT pg_catalog.to_jsonb(actual) =
               pg_catalog.to_jsonb(pg_catalog.jsonb_populate_record(
                   NULL::iam_one_time_tokens, p_record))
        INTO v_match FROM iam_one_time_tokens AS actual
        WHERE actual.id = (p_record->>'id')::uuid;
    ELSE
        RETURN false;
    END CASE;
    RETURN COALESCE(v_match, false);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_access_record_matches_v1(
    text, jsonb
) FROM PUBLIC;

