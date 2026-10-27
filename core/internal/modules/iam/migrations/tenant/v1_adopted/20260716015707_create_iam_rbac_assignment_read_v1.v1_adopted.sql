-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_read_v1(
    p_scope_schema name, p_id uuid,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (
    id uuid, resource_id uuid, target_kind text, target_id uuid,
    state text, revision bigint, created_at timestamptz,
    updated_at timestamptz, items jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_rbac_read_assert_v1(
        p_scope_schema, p_actor_id, p_required_permission_id
    ) THEN
        RETURN;
    END IF;
    RETURN QUERY
    SELECT assignment.id, assignment.resource_id, assignment.target_kind,
           assignment.target_id, assignment.state, assignment.revision,
           assignment.created_at, assignment.updated_at,
           (
               SELECT pg_catalog.jsonb_agg(
                   pg_catalog.jsonb_build_object(
                       'ordinal', member.ordinal,
                       'action_id', member.action_id,
                       'permission_id', member.permission_id,
                       'binding_id', member.binding_id
                   ) ORDER BY member.ordinal
               )
               FROM iam_rbac_assignment_members AS member
               WHERE member.group_id = assignment.id
                 AND member.group_revision = (
                     SELECT pg_catalog.max(history.group_revision)
                     FROM iam_rbac_assignment_members AS history
                     WHERE history.group_id = assignment.id
                 )
           ) AS items
    FROM iam_rbac_assignment_groups AS assignment
    WHERE assignment.id = p_id;
END;
$function$;

CREATE FUNCTION iam_rbac_assignment_list_v1(
    p_scope_schema name, p_state text, p_after_id uuid, p_limit integer,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (
    id uuid, resource_id uuid, target_kind text, target_id uuid,
    state text, revision bigint, created_at timestamptz,
    updated_at timestamptz, items jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 200 OR
       p_state NOT IN ('active', 'retired', 'all') THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment list input is invalid';
    END IF;
    IF NOT iam_rbac_read_assert_v1(
        p_scope_schema, p_actor_id, p_required_permission_id
    ) THEN
        RETURN;
    END IF;
    RETURN QUERY
    WITH page AS (
        SELECT assignment.*
        FROM iam_rbac_assignment_groups AS assignment
        WHERE (p_state = 'all' OR assignment.state = p_state)
          AND (p_after_id IS NULL OR assignment.id > p_after_id)
        ORDER BY assignment.id
        LIMIT p_limit
    )
    SELECT assignment.id, assignment.resource_id, assignment.target_kind,
           assignment.target_id, assignment.state, assignment.revision,
           assignment.created_at, assignment.updated_at,
           (
               SELECT pg_catalog.jsonb_agg(
                   pg_catalog.jsonb_build_object(
                       'ordinal', member.ordinal,
                       'action_id', member.action_id,
                       'permission_id', member.permission_id,
                       'binding_id', member.binding_id
                   ) ORDER BY member.ordinal
               )
               FROM iam_rbac_assignment_members AS member
               WHERE member.group_id = assignment.id
                 AND member.group_revision = (
                     SELECT pg_catalog.max(history.group_revision)
                     FROM iam_rbac_assignment_members AS history
                     WHERE history.group_id = assignment.id
                 )
           ) AS items
    FROM page AS assignment
    ORDER BY assignment.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_assignment_read_v1(
    name,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_assignment_list_v1(
    name,text,uuid,integer,uuid,uuid
) FROM PUBLIC;
