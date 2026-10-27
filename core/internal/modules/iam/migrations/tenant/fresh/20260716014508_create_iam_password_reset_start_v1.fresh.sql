-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_password_reset_start_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_token_id uuid, p_identifier_kind text,
    p_identifier_sha256 bytea, p_selector bytea, p_hash_key_id text,
    p_secret_hmac bytea, p_binding_sha256 bytea
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_identifier_id uuid;
    v_kind text;
    v_identifier_revision bigint;
    v_issue bigint;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF p_token_id IS NULL OR p_identifier_kind NOT IN ('email_v1', 'portal_v1') OR
       pg_catalog.octet_length(p_identifier_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.password_reset.start',
        p_request_sha256, p_token_id
    );
    IF FOUND THEN RETURN; END IF;
    SELECT identifier.user_id, identifier.id, identifier.kind,
        identifier.revision
    INTO v_user_id, v_identifier_id, v_kind, v_identifier_revision
    FROM iam_user_identifiers AS identifier
    JOIN iam_users AS subject ON subject.id = identifier.user_id
    WHERE identifier.kind = p_identifier_kind
      AND identifier.normalized_sha256 = p_identifier_sha256
      AND identifier.state = 'active' AND subject.status = 'active'
    FOR UPDATE OF identifier, subject;
    IF NOT FOUND THEN RETURN; END IF;
    v_issue := iam_one_time_issue_v1(
        p_token_id, 'password_reset_v1', v_user_id,
        v_identifier_id, v_kind, v_identifier_revision,
        NULL, NULL, p_selector, p_hash_key_id, p_secret_hmac,
        p_binding_sha256
    );
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.password_reset.start', p_request_sha256,
        p_token_id, v_user_id, 1, 'credential_target',
        v_user_id, 'one_time_token', false, false
    ) AS result;
    RETURN QUERY SELECT p_token_id, v_user_id, 1::bigint,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_reset_start_v1(
    name,uuid,uuid,bytea,uuid,text,bytea,bytea,text,bytea,bytea
) FROM PUBLIC;
