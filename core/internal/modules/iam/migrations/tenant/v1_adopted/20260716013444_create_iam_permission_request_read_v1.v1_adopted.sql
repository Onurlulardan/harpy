-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_read_v1(
    p_scope_schema name,
    p_actor_user_id uuid,
    p_group_id uuid
)
RETURNS TABLE (
    group_id uuid, requester_user_id uuid, request_mode text, group_state text,
    reviewer_user_id uuid, review_task_id uuid, group_revision bigint,
    expires_at timestamptz, terminal_at timestamptz, created_at timestamptz,
    updated_at timestamptz, request_documents jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request read scope is invalid';
    END IF;
    IF p_actor_user_id IS NULL OR p_group_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM iam_users AS actor
        WHERE actor.id OPERATOR(pg_catalog.=) p_actor_user_id
          AND actor.status OPERATOR(pg_catalog.=) 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request read authority is denied';
    END IF;
    RETURN QUERY
    SELECT request_group.id, request_group.requester_user_id,
           request_group.request_mode, request_group.state,
           request_group.reviewer_user_id, request_group.review_task_id,
           request_group.revision, request_group.expires_at,
           request_group.terminal_at, request_group.created_at,
           request_group.updated_at,
           pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'id', request.id, 'ordinal', request.ordinal,
               'target_kind', request.target_kind,
               'target_id', CASE request.target_kind
                   WHEN 'USER' THEN request.target_user_id
                   WHEN 'ROLE' THEN request.target_role_id
                   ELSE request.target_organization_id END,
               'permission_id', request.permission_id,
               'permission_protected', request.permission_is_protected,
               'planned_binding_id', request.planned_binding_id,
               'granted_binding_id', request.granted_binding_id,
               'state', request.state, 'revision', request.revision
           ) ORDER BY request.ordinal)
    FROM iam_permission_request_groups AS request_group
    JOIN iam_permission_requests AS request
      ON request.group_id OPERATOR(pg_catalog.=) request_group.id
    WHERE request_group.id OPERATOR(pg_catalog.=) p_group_id AND (
        request_group.requester_user_id OPERATOR(pg_catalog.=) p_actor_user_id OR
        request_group.reviewer_user_id OPERATOR(pg_catalog.=) p_actor_user_id OR
        iam_user_has_permission_v1(
            p_actor_user_id, request_group.reviewer_authority_permission_id
        )
    )
    GROUP BY request_group.id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM permission request is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_read_v1(name,uuid,uuid)
    FROM PUBLIC;
