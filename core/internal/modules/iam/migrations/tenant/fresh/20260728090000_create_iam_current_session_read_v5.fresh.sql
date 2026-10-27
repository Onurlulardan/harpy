-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_current_session_read_v5(
    p_scope_schema name, p_user_id uuid, p_session_id uuid
)
RETURNS TABLE (
    result_user_id uuid, result_session_id uuid, result_client_id text,
    result_access_mode text, result_portal_binding_id uuid,
    result_token_version bigint, result_authorization_version bigint,
    result_deployment_mode text, result_tenant_purpose text,
    result_platform_subject_id uuid, result_platform_clock bigint,
    result_security_epoch bytea, result_tenant_capabilities text[],
    result_password_credential_revision bigint, result_email text,
    result_display_name text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT session_view.result_user_id, session_view.result_session_id,
           session_view.result_client_id, session_view.result_access_mode,
           session_view.result_portal_binding_id,
           session_view.result_token_version,
           session_view.result_authorization_version,
           session_view.result_deployment_mode,
           session_view.result_tenant_purpose,
           session_view.result_platform_subject_id,
           session_view.result_platform_clock,
           session_view.result_security_epoch,
           ARRAY(
               SELECT capability
               FROM pg_catalog.unnest(
                   session_view.result_tenant_capabilities
               ) AS capability
               ORDER BY capability COLLATE "C"
           ),
           session_view.result_password_credential_revision,
           session_view.result_email,
           session_view.result_display_name
    FROM iam_current_session_read_v4(p_scope_schema, p_user_id, p_session_id)
         AS session_view;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_current_session_read_v5(name,uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v5(name,uuid,uuid)
TO forgecrud_runtime;
