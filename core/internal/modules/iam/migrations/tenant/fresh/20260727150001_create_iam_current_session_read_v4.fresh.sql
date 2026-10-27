-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_current_session_read_v4(
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
           session_view.result_tenant_capabilities,
           session_view.result_password_credential_revision,
           identifier.display_value,
           COALESCE(
               NULLIF(pg_catalog.btrim(subject.profile_values->>'display_name'), ''),
               identifier.display_value
           )
    FROM iam_current_session_read_v3(p_scope_schema, p_user_id, p_session_id)
         AS session_view
    JOIN iam_users AS subject
      ON subject.id OPERATOR(pg_catalog.=) session_view.result_user_id
     AND subject.status OPERATOR(pg_catalog.=) 'active'
    JOIN iam_user_identifiers AS identifier
      ON identifier.user_id OPERATOR(pg_catalog.=) session_view.result_user_id
     AND identifier.kind OPERATOR(pg_catalog.=) 'email_v1'
     AND identifier.state OPERATOR(pg_catalog.=) 'active';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_current_session_read_v4(name,uuid,uuid) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v3(name,uuid,uuid)
FROM forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)
TO forgecrud_v2_runtime;
