-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_actor_assert_v1(
    p_scope_schema name, p_actor_id uuid, p_required_permission_id uuid,
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
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC command scope is invalid';
    END IF;
    IF p_actor_id IS NULL OR p_required_permission_id IS NULL OR
       p_expected_token_version < 1 OR p_expected_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC command evidence is invalid';
    END IF;
    SELECT status, token_version, authorization_version
    INTO v_status, v_token_version, v_authorization_version
    FROM iam_users WHERE id = p_actor_id FOR UPDATE;
    IF NOT FOUND OR v_status IS DISTINCT FROM 'active' OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version OR
       NOT iam_user_has_permission_v1(p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC command authority is denied';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_actor_assert_v1(
    name,uuid,uuid,bigint,bigint
) FROM PUBLIC;
