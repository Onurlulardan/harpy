-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_permission_request_authority_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.id = NEW.creator_authority_permission_id
          AND permission.match_kind = 'exact'
          AND permission.is_protected
          AND permission.state = 'active'
          AND resource.resource_key = 'iam.permission_requests'
          AND resource.state = 'active'
          AND action.action_key = 'manage'
          AND action.state = 'active'
    ) OR NOT EXISTS (
        SELECT 1
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.id = NEW.reviewer_authority_permission_id
          AND permission.match_kind = 'exact'
          AND permission.is_protected
          AND permission.state = 'active'
          AND resource.resource_key = 'iam.permission_requests'
          AND resource.state = 'active'
          AND action.action_key = 'review'
          AND action.state = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_permission_request_groups_authority_guard',
            MESSAGE = 'IAM permission request authority catalog is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_permission_request_authority_v1()
    FROM PUBLIC;

CREATE TRIGGER iam_permission_request_groups_authority_guard
BEFORE INSERT ON iam_permission_request_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_request_authority_v1();
ALTER TABLE iam_permission_request_groups ENABLE ALWAYS TRIGGER
    iam_permission_request_groups_authority_guard;
