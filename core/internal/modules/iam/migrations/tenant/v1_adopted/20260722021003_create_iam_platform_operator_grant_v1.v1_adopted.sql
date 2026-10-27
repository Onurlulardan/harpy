-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_platform_operator_grant_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint, p_expected_platform_clock bigint,
    p_purpose_sha256 bytea, p_auth_time timestamptz, p_acr text,
    p_amr_password boolean, p_amr_webauthn boolean, p_amr_totp boolean,
    p_target_user_id uuid, p_target_expected_authorization_version bigint,
    p_subject_id uuid, p_role_id uuid, p_subject_role_binding_id uuid,
    p_role_permission_binding_id uuid, p_capability text
)
RETURNS TABLE (
    target_user_id uuid, subject_id uuid, role_id uuid,
    subject_role_binding_id uuid, role_permission_binding_id uuid,
    target_authorization_version bigint, platform_clock bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_target_current_version bigint;
    v_subject_id uuid;
    v_role_id uuid;
    v_subject_role_id uuid;
    v_role_permission_id uuid;
    v_target_result_version bigint;
    v_platform_clock bigint;
    v_replayed boolean;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_target_user_id IS NULL OR p_actor_id = p_target_user_id OR
       p_target_expected_authorization_version < 1 OR p_subject_id IS NULL OR
       p_role_id IS NULL OR p_subject_role_binding_id IS NULL OR
       p_role_permission_binding_id IS NULL OR p_capability !~ '^platform\.' OR
       pg_catalog.octet_length(p_purpose_sha256) <> 32 OR
       p_acr <> 'urn:forgecrud:acr:step-up' OR
       NOT (p_amr_password OR p_amr_webauthn OR p_amr_totp) OR
       p_auth_time > v_now OR p_auth_time < v_now - INTERVAL '10 minutes' OR
       p_purpose_sha256 <> pg_catalog.sha256(pg_catalog.convert_to(
           'platform.operators.grant' || E'\n' || p_target_user_id::text ||
           E'\n' || p_capability, 'UTF8')) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Platform operator grant input is invalid';
    END IF;
    SELECT identity.tenant_id INTO v_tenant_id
    FROM tenancy_schema_identity AS identity WHERE identity.singleton;
    PERFORM 1 FROM iam_users AS principal
    WHERE principal.id = p_actor_id AND principal.status = 'active'
      AND principal.identity_kind = 'human'
      AND principal.token_version = p_expected_token_version
      AND principal.authorization_version = p_expected_authorization_version
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Platform owner identity evidence is stale';
    END IF;
    SELECT principal.authorization_version INTO v_target_current_version
    FROM iam_users AS principal
    WHERE principal.id = p_target_user_id AND principal.status = 'active'
      AND principal.identity_kind = 'human'
      AND principal.authorization_version IN (
          p_target_expected_authorization_version,
          p_target_expected_authorization_version + 1
      )
    FOR UPDATE;
    IF v_tenant_id IS NULL OR v_target_current_version IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P1404',
            MESSAGE = 'Platform operator target is not an active primary user';
    END IF;

    SELECT result.subject_id, result.role_id, result.subject_role_binding_id,
           result.role_permission_binding_id, result.target_authorization_version,
           result.platform_clock, result.replayed
    INTO v_subject_id, v_role_id, v_subject_role_id, v_role_permission_id,
         v_target_result_version, v_platform_clock, v_replayed
    FROM platform.iam_platform_operator_grant_apply_v1(
        v_tenant_id, p_operation_id, p_request_sha256, p_actor_id,
        p_target_user_id, p_target_expected_authorization_version + 1,
        p_expected_platform_clock, p_subject_id, p_role_id,
        p_subject_role_binding_id, p_role_permission_binding_id, p_capability
    ) AS result;

    IF v_replayed THEN
        IF v_target_current_version <> v_target_result_version THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'Platform operator grant replay identity is stale';
        END IF;
    ELSE
        IF v_target_current_version <> p_target_expected_authorization_version THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'Platform operator grant target changed';
        END IF;
        UPDATE iam_users SET authorization_version = v_target_result_version,
            revision = revision + 1, updated_at = v_now
        WHERE id = p_target_user_id
          AND authorization_version = p_target_expected_authorization_version;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'Platform operator grant target changed';
        END IF;
    END IF;
    RETURN QUERY SELECT p_target_user_id, v_subject_id, v_role_id,
        v_subject_role_id, v_role_permission_id, v_target_result_version,
        v_platform_clock, v_replayed;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_platform_operator_grant_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bigint,uuid,uuid,uuid,uuid,text
) FROM PUBLIC;
