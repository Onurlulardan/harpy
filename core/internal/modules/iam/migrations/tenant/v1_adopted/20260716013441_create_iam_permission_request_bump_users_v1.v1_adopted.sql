-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_bump_users_v1(
    p_group_id uuid,
    p_now timestamptz
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_count integer;
BEGIN
    UPDATE iam_users AS principal
    SET authorization_version = principal.authorization_version + 1,
        revision = principal.revision + 1,
        updated_at = p_now
    WHERE principal.status = 'active' AND (
        EXISTS (
            SELECT 1 FROM iam_permission_requests AS request
            WHERE request.group_id = p_group_id AND request.target_kind = 'USER'
              AND request.target_user_id = principal.id
        ) OR EXISTS (
            SELECT 1
            FROM iam_permission_requests AS request
            JOIN iam_user_role_assignments AS assignment
              ON assignment.role_id = request.target_role_id
            WHERE request.group_id = p_group_id AND request.target_kind = 'ROLE'
              AND assignment.user_id = principal.id AND assignment.state = 'active'
        ) OR EXISTS (
            SELECT 1
            FROM iam_permission_requests AS request
            JOIN iam_organization_memberships AS membership
              ON membership.organization_id = request.target_organization_id
            WHERE request.group_id = p_group_id
              AND request.target_kind = 'ORGANIZATION'
              AND membership.user_id = principal.id AND membership.state = 'active'
        )
    );
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_bump_users_v1(uuid,timestamptz)
    FROM PUBLIC;
