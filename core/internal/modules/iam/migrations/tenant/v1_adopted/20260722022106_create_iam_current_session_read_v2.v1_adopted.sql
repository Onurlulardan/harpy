-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_current_session_read_v2(
    p_scope_schema name, p_user_id uuid, p_session_id uuid
)
RETURNS TABLE (
    result_user_id uuid, result_session_id uuid, result_client_id text,
    result_access_mode text, result_portal_binding_id uuid,
    result_token_version bigint, result_authorization_version bigint,
    result_deployment_mode text, result_tenant_purpose text,
    result_platform_subject_id uuid, result_platform_clock bigint,
    result_security_epoch bytea, result_tenant_capabilities text[]
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
           context.deployment_mode, context.tenant_purpose, context.subject_id,
           context.platform_clock, session.security_epoch,
           session_view.result_permissions
    FROM iam_current_session_read_v1(p_scope_schema, p_user_id, p_session_id)
         AS session_view
    JOIN iam_sessions AS session ON session.id = session_view.result_session_id
    CROSS JOIN iam_platform_session_context_v1(
        p_scope_schema, session_view.result_user_id,
        session_view.result_token_version,
        session_view.result_authorization_version
    ) AS context
    WHERE session.security_epoch = iam_security_epoch_current_v1(p_scope_schema)
      AND NOT EXISTS (
          SELECT 1
          FROM pg_catalog.unnest(session_view.result_permissions) AS capability(key)
          WHERE capability.key OPERATOR(pg_catalog.~) '^platform\.'
      );
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_current_session_read_v2(name,uuid,uuid) FROM PUBLIC;
