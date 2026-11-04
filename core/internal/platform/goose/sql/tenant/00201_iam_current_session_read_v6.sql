-- +goose Up
CREATE FUNCTION iam_current_session_read_v6(
    in_schema_name name,
    in_user_id uuid,
    in_session_id uuid
)
RETURNS TABLE (
    result_user_id uuid,
    result_session_id uuid,
    result_client_id text,
    result_access_mode text,
    result_portal_binding_id uuid,
    result_token_version bigint,
    result_authorization_version bigint,
    result_deployment_mode text,
    result_tenant_purpose text,
    result_platform_subject_id uuid,
    result_platform_clock bigint,
    result_security_epoch bytea,
    result_tenant_capabilities text[],
    result_password_credential_revision bigint,
    result_email text,
    result_display_name text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_user_id,
        in_session_id,
        ''::text,
        ''::text,
        NULL::uuid,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        NULL::uuid,
        0::bigint,
        ''::bytea,
        '{}'::text[],
        NULL::bigint,
        ''::text,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION iam_current_session_read_v6(name, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_current_session_read_v6(name, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_current_session_read_v6(name, uuid, uuid);
