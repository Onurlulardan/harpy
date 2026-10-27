-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauth_binding_sha256_v1(
    p_tenant_id uuid, p_user_id uuid, p_session_id uuid,
    p_token_version bigint, p_authorization_version bigint,
    p_access_mode text, p_portal_binding_id uuid
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_tenant_id IS NULL OR p_user_id IS NULL OR p_session_id IS NULL OR
       p_token_version OPERATOR(pg_catalog.<=) 0 OR
       p_authorization_version OPERATOR(pg_catalog.<=) 0 OR NOT (
           (p_access_mode IS NOT DISTINCT FROM 'operator_console' AND
            p_portal_binding_id IS NULL) OR
           (p_access_mode IS NOT DISTINCT FROM 'portal' AND
            p_portal_binding_id IS NOT NULL)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM reauthentication binding input is invalid';
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.convert_to('harpy:iam-reauth-session:v1', 'UTF8')
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_tenant_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_user_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_session_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.int8send(p_token_version)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.int8send(p_authorization_version)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.convert_to(p_access_mode, 'UTF8')
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||)
        pg_catalog.uuid_send(COALESCE(
            p_portal_binding_id,
            '00000000-0000-0000-0000-000000000000'::uuid
        )) OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
    );
END;
$function$;

CREATE FUNCTION iam_reauth_context_sha256_v1(
    p_tenant_id uuid, p_user_id uuid, p_session_id uuid,
    p_purpose_sha256 bytea, p_action_policy_version bigint,
    p_method text, p_origin_sha256 bytea
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_tenant_id IS NULL OR p_user_id IS NULL OR p_session_id IS NULL OR
       pg_catalog.octet_length(p_purpose_sha256) IS DISTINCT FROM 32 OR
       p_action_policy_version OPERATOR(pg_catalog.<=) 0 OR
       p_method NOT IN ('password_v1', 'webauthn_v1', 'totp_v1') OR
       pg_catalog.octet_length(p_origin_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM reauthentication context input is invalid';
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.convert_to('harpy:iam-reauth-context:v1', 'UTF8')
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_tenant_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_user_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_session_id)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) p_purpose_sha256
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.int8send(p_action_policy_version)
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) pg_catalog.convert_to(p_method, 'UTF8')
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
        OPERATOR(pg_catalog.||) p_origin_sha256
        OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauth_binding_sha256_v1(
    uuid,uuid,uuid,bigint,bigint,text,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_reauth_context_sha256_v1(
    uuid,uuid,uuid,bytea,bigint,text,bytea
) FROM PUBLIC;
