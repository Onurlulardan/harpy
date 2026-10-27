-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_self_profile_read_v1(
    p_scope_schema name,
    p_actor_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    email text,
    status text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    roles jsonb,
    organizations jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL OR p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM self profile read scope is invalid';
    END IF;
    PERFORM 1 FROM iam_users AS actor
    WHERE actor.id = p_actor_id
      AND actor.status = 'active'
      AND actor.token_version = p_expected_token_version
      AND actor.authorization_version = p_expected_authorization_version;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM self profile read authority is denied';
    END IF;
    RETURN QUERY
    SELECT actor.id, identifier.display_value, actor.status, actor.revision,
           actor.created_at, actor.updated_at,
           COALESCE((
               SELECT pg_catalog.jsonb_agg(
                   pg_catalog.jsonb_build_object(
                       'id', role.id, 'name', role.name
                   ) ORDER BY role.id
               )
               FROM iam_user_role_assignments AS assignment
               JOIN iam_roles AS role ON role.id = assignment.role_id
               WHERE assignment.user_id = actor.id
                 AND assignment.state = 'active'
                 AND role.status = 'active'
           ), '[]'::jsonb),
           COALESCE((
               SELECT pg_catalog.jsonb_agg(
                   pg_catalog.jsonb_build_object(
                       'id', organization.id,
                       'name', organization.name,
                       'slug', organization.slug
                   ) ORDER BY organization.id
               )
               FROM iam_organization_memberships AS membership
               JOIN iam_organizations AS organization
                 ON organization.id = membership.organization_id
               WHERE membership.user_id = actor.id
                 AND membership.state = 'active'
                 AND organization.status = 'active'
           ), '[]'::jsonb)
    FROM iam_users AS actor
    JOIN iam_user_identifiers AS identifier
      ON identifier.user_id = actor.id
     AND identifier.kind = 'email_v1'
     AND identifier.state = 'active'
    WHERE actor.id = p_actor_id
      AND actor.status = 'active'
      AND actor.token_version = p_expected_token_version
      AND actor.authorization_version = p_expected_authorization_version;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5404',
            MESSAGE = 'IAM self profile is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_self_profile_read_v1(
    name,uuid,bigint,bigint
) FROM PUBLIC;
