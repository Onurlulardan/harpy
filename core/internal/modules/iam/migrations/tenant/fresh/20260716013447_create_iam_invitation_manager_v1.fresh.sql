-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_manager_v1(p_user_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_permission_id uuid;
BEGIN
    SELECT permission.id
    INTO v_permission_id
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE resource.resource_key OPERATOR(pg_catalog.=) 'iam.invitations'
      AND resource.state OPERATOR(pg_catalog.=) 'active'
      AND action.action_key OPERATOR(pg_catalog.=) 'manage'
      AND action.state OPERATOR(pg_catalog.=) 'active'
      AND permission.match_kind OPERATOR(pg_catalog.=) 'exact'
      AND permission.state OPERATOR(pg_catalog.=) 'active';
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    RETURN iam_user_has_permission_v1(p_user_id, v_permission_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_manager_v1(uuid) FROM PUBLIC;
