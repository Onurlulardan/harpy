-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION iam_password_reset_complete_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_selector bytea, p_hash_key_id text,
    p_secret_hmac bytea, p_binding_sha256 bytea,
    p_salt bytea, p_tag bytea
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_token_id uuid; v_user_id uuid; v_identifier_id uuid;
    v_identifier_revision bigint; v_token_state text;
    v_expires_at timestamptz; v_credential_revision bigint;
    v_token_version bigint; v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       NOT iam_hmac_key_id_valid_v1(p_hash_key_id) OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_binding_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_salt) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_tag) IS DISTINCT FROM 32 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.password_reset.complete', p_request_sha256, NULL
    );
    IF FOUND THEN RETURN; END IF;
    SELECT token.id, token.user_id, token.identifier_id,
        token.identifier_revision, token.state, token.expires_at
    INTO v_token_id, v_user_id, v_identifier_id,
        v_identifier_revision, v_token_state, v_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.purpose = 'password_reset_v1' AND token.selector = p_selector
      AND token.hash_key_id = p_hash_key_id AND token.secret_hmac = p_secret_hmac
      AND token.binding_sha256 = p_binding_sha256 FOR UPDATE;
    IF NOT FOUND OR v_token_state <> 'pending' OR v_expires_at <= v_now OR
       NOT EXISTS (
          SELECT 1 FROM iam_user_identifiers AS identifier
          WHERE identifier.id = v_identifier_id
            AND identifier.user_id = v_user_id
            AND identifier.revision = v_identifier_revision
            AND identifier.state = 'active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '28000',
            MESSAGE = 'IAM password reset proof is invalid';
    END IF;
    INSERT INTO iam_password_credentials (
        user_id, credential_version, state, algorithm, salt, tag,
        memory_kib, iterations, parallelism, upgrade_required,
        revision, created_at, updated_at
    ) VALUES (
        v_user_id, 1, 'active', 'argon2id_v1', p_salt, p_tag,
        65536, 3, 4, false, 1, v_now, v_now
    ) ON CONFLICT (user_id) DO UPDATE SET
        credential_version = iam_password_credentials.credential_version + 1,
        state = 'active', algorithm = 'argon2id_v1', salt = EXCLUDED.salt,
        tag = EXCLUDED.tag, memory_kib = 65536, iterations = 3,
        parallelism = 4, legacy_bcrypt_hash = NULL, upgrade_required = false,
        revision = iam_password_credentials.revision + 1,
        updated_at = v_now, invalidated_at = NULL
    RETURNING revision INTO v_credential_revision;
    UPDATE iam_one_time_tokens SET state = 'consumed', consumed_at = v_now,
        revision = revision + 1, updated_at = v_now WHERE id = v_token_id;
    UPDATE iam_one_time_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = v_user_id AND state = 'pending';
    UPDATE iam_sessions SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = v_user_id AND state = 'active';
    UPDATE iam_refresh_families SET state = 'revoked',
        revoke_reason = 'password_reset', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = v_user_id AND state = 'active';
    UPDATE iam_refresh_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE family_id IN (SELECT id FROM iam_refresh_families
        WHERE user_id = v_user_id) AND state = 'active';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.password_reset.complete', p_request_sha256,
        v_user_id, v_user_id, v_credential_revision,
        'credential_target', v_user_id, 'password_credential', true, true
    ) AS result;
    RETURN QUERY SELECT v_user_id, v_user_id, v_credential_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_reset_complete_v1(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea,bytea,bytea
) FROM PUBLIC;
