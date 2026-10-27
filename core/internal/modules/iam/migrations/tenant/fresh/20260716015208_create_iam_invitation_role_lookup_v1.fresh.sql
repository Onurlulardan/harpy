-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_role_lookup_v1(
    p_scope_schema name, p_actor_user_id uuid, p_role_name text
)
RETURNS TABLE (
    role_id uuid, role_name text, role_scope_kind text,
    role_scope_id uuid, organization_id uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_user_id IS NULL OR
       pg_catalog.char_length(p_role_name) NOT BETWEEN 1 AND 255 OR
       NOT EXISTS (
           SELECT 1 FROM iam_users AS actor
           WHERE actor.id = p_actor_user_id AND actor.status = 'active'
       ) OR NOT iam_invitation_manager_v1(p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation role lookup is denied';
    END IF;
    RETURN QUERY
    SELECT role.id, role.name, role.scope_kind, role.scope_id,
           role.organization_id
    FROM iam_roles AS role
    WHERE role.name_key = pg_catalog.convert_to(p_role_name, 'UTF8')
      AND role.status = 'active' AND NOT role.is_protected;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM invitation role is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_role_lookup_v1(name,uuid,text) FROM PUBLIC;

