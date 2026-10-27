-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identifier_verify_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_selector bytea, p_hash_key_id text,
    p_secret_hmac bytea, p_binding_sha256 bytea
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_token_id uuid;
    v_user_id uuid;
    v_identifier_id uuid;
    v_identifier_revision bigint;
    v_token_state text;
    v_expires_at timestamptz;
    v_identifier_state text;
    v_result_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_binding_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM identifier verify input is invalid';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.identifier.verify', p_request_sha256, NULL
    );
    IF FOUND THEN RETURN; END IF;
    SELECT token.id, token.user_id, token.identifier_id,
        token.identifier_revision, token.state, token.expires_at
    INTO v_token_id, v_user_id, v_identifier_id,
        v_identifier_revision, v_token_state, v_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.purpose = 'email_verification_v1'
      AND token.selector = p_selector AND token.hash_key_id = p_hash_key_id
      AND token.secret_hmac = p_secret_hmac
      AND token.binding_sha256 = p_binding_sha256
    FOR UPDATE;
    IF NOT FOUND OR v_token_state <> 'pending' OR v_expires_at <= v_now THEN
        RAISE EXCEPTION USING ERRCODE = '28000',
            MESSAGE = 'IAM email verification proof is invalid';
    END IF;
    SELECT identifier.state, identifier.revision
    INTO v_identifier_state, v_result_revision
    FROM iam_user_identifiers AS identifier
    WHERE identifier.id = v_identifier_id
      AND identifier.user_id = v_user_id
      AND identifier.kind = 'email_v1'
      AND identifier.revision = v_identifier_revision
    FOR UPDATE;
    IF NOT FOUND OR v_identifier_state NOT IN ('pending', 'verified') THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM email identifier changed';
    END IF;
    UPDATE iam_user_identifiers SET state = 'active',
        verified_at = COALESCE(verified_at, v_now), activated_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = v_identifier_id RETURNING revision INTO v_result_revision;
    UPDATE iam_one_time_tokens SET state = 'consumed', consumed_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = v_token_id;
    UPDATE iam_one_time_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE purpose = 'email_verification_v1'
      AND user_id = v_user_id AND identifier_id = v_identifier_id
      AND state = 'pending';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.identifier.verify', p_request_sha256,
        v_identifier_id, v_user_id, v_result_revision,
        'credential_target', v_user_id, 'identifier', true, true
    ) AS result;
    RETURN QUERY SELECT v_identifier_id, v_user_id, v_result_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identifier_verify_v1(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea
) FROM PUBLIC;
