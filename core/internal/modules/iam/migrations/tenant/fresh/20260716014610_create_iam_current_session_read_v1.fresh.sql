-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_current_session_read_v1(
    p_scope_schema name,
    p_user_id uuid,
    p_session_id uuid
)
RETURNS TABLE (
    result_user_id uuid,
    result_session_id uuid,
    result_client_id text,
    result_access_mode text,
    result_portal_binding_id uuid,
    result_token_version bigint,
    result_authorization_version bigint,
    result_permissions text[]
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_user_id IS NULL OR p_session_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM current session caller is invalid';
    END IF;
    RETURN QUERY
    SELECT principal.id, session.id,
           CASE session.access_mode
               WHEN 'operator_console' THEN 'forgecrud-operator-console-v1'
               WHEN 'portal' THEN 'forgecrud-portal-v1'
           END,
           session.access_mode, session.portal_binding_id,
           principal.token_version, principal.authorization_version,
           ARRAY(
               SELECT resource.resource_key || '.' || action.action_key
               FROM iam_permissions AS permission
               JOIN iam_resources AS resource ON resource.id = permission.resource_id
               JOIN iam_actions AS action ON action.id = permission.action_id
               WHERE permission.match_kind = 'exact'
                 AND permission.state = 'active'
                 AND resource.state = 'active' AND action.state = 'active'
                 AND iam_user_has_permission_v1(principal.id, permission.id)
               ORDER BY resource.resource_key, action.action_key
           )
    FROM iam_users AS principal
    JOIN iam_sessions AS session ON session.user_id = principal.id
    JOIN iam_user_access_modes AS access
      ON access.id = session.access_mode_id
     AND access.user_id = principal.id AND access.mode = session.access_mode
    LEFT JOIN iam_portal_subject_bindings AS portal
      ON portal.id = session.portal_binding_id
     AND portal.user_id = principal.id AND portal.access_mode_id = access.id
    WHERE principal.id = p_user_id AND session.id = p_session_id
      AND principal.status = 'active' AND principal.identity_kind = 'human'
      AND session.state = 'active'
      AND session.absolute_expires_at > pg_catalog.statement_timestamp()
      AND access.state = 'active'
      AND (
          (session.access_mode = 'operator_console' AND portal.id IS NULL) OR
          (session.access_mode = 'portal' AND portal.state = 'active')
      );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_current_session_read_v1(name,uuid,uuid) FROM PUBLIC;
