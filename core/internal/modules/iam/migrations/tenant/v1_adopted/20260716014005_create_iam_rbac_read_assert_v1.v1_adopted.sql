-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_read_assert_v1(
    p_scope_schema name, p_actor_id uuid, p_required_permission_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC query scope is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_users WHERE id = p_actor_id AND status = 'active'
    ) OR NOT iam_user_has_permission_v1(p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC query authority is denied';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_read_assert_v1(name,uuid,uuid) FROM PUBLIC;
