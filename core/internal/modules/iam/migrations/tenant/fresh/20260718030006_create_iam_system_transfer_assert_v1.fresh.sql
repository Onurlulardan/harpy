-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_transfer_assert_v1(
    p_scope_schema name,p_actor_id uuid,p_resource_kind text,p_action text,
    p_token_version bigint,p_authorization_version bigint,p_command boolean
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_permission_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('users','roles','organizations') OR
       p_action NOT IN ('import','export') OR p_command IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM system transfer authority is invalid';
    END IF;
    SELECT permission.id INTO STRICT v_permission_id
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id=permission.resource_id
    JOIN iam_actions AS action ON action.id=permission.action_id
    WHERE resource.resource_key='iam.'||p_resource_kind
      AND resource.owner_domain='iam' AND resource.state='active'
      AND action.action_key=p_action AND action.state='active'
      AND permission.match_kind='exact' AND permission.is_protected
      AND permission.state='active';
    IF p_command THEN
        RETURN iam_identity_actor_assert_v1(
            p_scope_schema,p_actor_id,v_permission_id,
            p_token_version,p_authorization_version
        );
    END IF;
    RETURN iam_identity_query_assert_v1(
        p_scope_schema,p_actor_id,v_permission_id,
        p_token_version,p_authorization_version
    );
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='IAM system transfer authority is unavailable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_transfer_assert_v1(
    name,uuid,text,text,bigint,bigint,boolean
) FROM PUBLIC;

