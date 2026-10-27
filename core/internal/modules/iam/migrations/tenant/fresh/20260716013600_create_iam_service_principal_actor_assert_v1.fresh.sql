-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_service_principal_actor_assert_v1(
    p_scope_schema name,
    p_actor_id uuid,
    p_manage_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_purpose_sha256 bytea,
    p_auth_time timestamptz,
    p_amr_password boolean,
    p_amr_webauthn boolean,
    p_amr_totp boolean
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz;
    v_state text;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM service principal command scope is invalid';
    END IF;
    IF p_actor_id IS NULL OR p_manage_permission_id IS NULL OR
       p_expected_token_version < 1 OR p_expected_authorization_version < 1 OR
       pg_catalog.octet_length(p_purpose_sha256) IS DISTINCT FROM 32 OR
       p_auth_time IS NULL OR p_auth_time > v_now OR
       p_auth_time < v_now - INTERVAL '10 minutes' OR
       NOT (p_amr_password OR p_amr_webauthn OR p_amr_totp) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM service principal command evidence is invalid';
    END IF;
    SELECT actor.status, actor.token_version, actor.authorization_version
    INTO v_state, v_token_version, v_authorization_version
    FROM iam_users AS actor
    WHERE actor.id OPERATOR(pg_catalog.=) p_actor_id
    FOR UPDATE;
    IF NOT FOUND OR v_state IS DISTINCT FROM 'active' OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version OR
       NOT iam_user_has_permission_v1(p_actor_id, p_manage_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM service principal command authority is denied';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_service_principal_actor_assert_v1(
    name,uuid,uuid,bigint,bigint,bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;
