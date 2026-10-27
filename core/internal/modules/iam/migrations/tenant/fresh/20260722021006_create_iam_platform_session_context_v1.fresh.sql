-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_platform_session_context_v1(
    p_scope_schema name, p_actor_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    deployment_mode text, tenant_purpose text, subject_id uuid,
    platform_clock bigint, authorization_version bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_auth_version bigint;
    v_token_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL OR p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Platform session context input is invalid';
    END IF;
    SELECT identity.tenant_id, principal.authorization_version,
           principal.token_version
    INTO v_tenant_id, v_auth_version, v_token_version
    FROM tenancy_schema_identity AS identity
    JOIN iam_users AS principal ON principal.id = p_actor_id
      AND principal.status = 'active' AND principal.identity_kind = 'human'
    WHERE identity.singleton;
    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Platform session context is unavailable';
    END IF;
    IF v_token_version <> p_expected_token_version OR
       v_auth_version <> p_expected_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Platform session identity evidence is stale';
    END IF;
    RETURN QUERY
    SELECT context.deployment_mode, context.tenant_purpose,
           context.subject_id, context.platform_clock, v_auth_version
    FROM platform.iam_platform_session_context_read_v1(
        v_tenant_id, p_actor_id
    ) AS context;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_platform_session_context_v1(name,uuid,bigint,bigint)
FROM PUBLIC;
