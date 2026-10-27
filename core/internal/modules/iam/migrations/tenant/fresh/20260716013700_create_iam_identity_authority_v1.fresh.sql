-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_actor_assert_v1(
    p_scope_schema name, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_status text;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL OR p_permission_id IS NULL OR
       p_expected_token_version < 1 OR p_expected_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity command scope is invalid';
    END IF;
    SELECT status, token_version, authorization_version
    INTO v_status, v_token_version, v_authorization_version
    FROM iam_users WHERE id = p_actor_id FOR UPDATE;
    IF NOT FOUND OR v_status IS DISTINCT FROM 'active' OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version OR
       NOT iam_user_has_permission_v1(p_actor_id, p_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity command authority is denied';
    END IF;
    RETURN true;
END;
$function$;

CREATE FUNCTION iam_identity_query_assert_v1(
    p_scope_schema name, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL OR p_permission_id IS NULL OR
       NOT EXISTS (
          SELECT 1 FROM iam_users AS actor
          WHERE actor.id = p_actor_id AND actor.status = 'active'
            AND actor.token_version = p_expected_token_version
            AND actor.authorization_version = p_expected_authorization_version
       ) OR NOT iam_user_has_permission_v1(p_actor_id, p_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity query authority is denied';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_actor_assert_v1(
    name,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_identity_query_assert_v1(
    name,uuid,uuid,bigint,bigint
) FROM PUBLIC;
