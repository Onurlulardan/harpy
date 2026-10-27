-- +goose Up
-- Migration stream version: 20260722140103.
-- +goose StatementBegin
CREATE FUNCTION iam_initial_admin_setup_accept_v1(
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
    v_setup iam_initial_admin_setups%ROWTYPE;
    v_user_id uuid;
    v_identifier_id uuid;
    v_token_state text;
    v_expires_at timestamptz;
    v_credential_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_tenant_id IS NULL OR p_operation_id IS NULL
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR pg_catalog.octet_length(p_selector) <> 16
       OR p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$'
       OR pg_catalog.octet_length(p_secret_hmac) <> 32
       OR pg_catalog.octet_length(p_binding_sha256) <> 32
       OR pg_catalog.octet_length(p_salt) <> 16
       OR pg_catalog.octet_length(p_tag) <> 32
       OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id=p_tenant_id
             AND tenant.schema_name=p_scope_schema::text
             AND tenant.status='active' AND tenant.deleted_at IS NULL
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id,'iam.initial_admin_setup.accept',p_request_sha256,NULL
    );
    IF FOUND THEN RETURN; END IF;
    SELECT setup.* INTO v_setup
    FROM iam_initial_admin_setups AS setup
    JOIN iam_one_time_tokens AS token ON token.id=setup.token_id
    WHERE token.purpose='initial_admin_setup.v1'
      AND token.selector=p_selector AND token.hash_key_id=p_hash_key_id
      AND token.secret_hmac=p_secret_hmac
      AND token.binding_sha256=p_binding_sha256
    FOR UPDATE OF setup,token;
    IF NOT FOUND OR v_setup.state<>'pending' THEN
        RAISE EXCEPTION USING ERRCODE='28000',
            MESSAGE='initial admin setup proof is invalid';
    END IF;
    SELECT token.user_id,token.identifier_id,token.state,token.expires_at
    INTO v_user_id,v_identifier_id,v_token_state,v_expires_at
    FROM iam_one_time_tokens AS token WHERE token.id=v_setup.token_id;
    IF v_token_state<>'pending' OR v_expires_at<=v_now
       OR v_user_id IS DISTINCT FROM v_setup.user_id
       OR EXISTS (
           SELECT 1 FROM iam_password_credentials AS credential
           WHERE credential.user_id=v_setup.user_id
       ) OR NOT EXISTS (
           SELECT 1 FROM iam_users AS app_user
           JOIN iam_user_identifiers AS identifier
             ON identifier.id=v_identifier_id
            AND identifier.user_id=app_user.id
           WHERE app_user.id=v_setup.user_id AND app_user.status='active'
             AND app_user.identity_kind='human'
             AND identifier.kind='email_v1' AND identifier.state='active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='28000',
            MESSAGE='initial admin setup target is invalid';
    END IF;
    INSERT INTO iam_password_credentials (
        user_id,credential_version,state,algorithm,salt,tag,memory_kib,
        iterations,parallelism,upgrade_required,revision,created_at,updated_at
    ) VALUES (
        v_setup.user_id,1,'active','argon2id_v1',p_salt,p_tag,65536,
        3,4,false,1,v_now,v_now
    ) RETURNING revision INTO v_credential_revision;
    UPDATE iam_one_time_tokens SET state='consumed',consumed_at=v_now,
        revision=revision+1,updated_at=v_now WHERE id=v_setup.token_id;
    UPDATE iam_one_time_tokens SET state='revoked',revoked_at=v_now,
        revision=revision+1,updated_at=v_now
    WHERE user_id=v_setup.user_id AND state='pending';
    UPDATE iam_initial_admin_setups SET state='accepted',accepted_at=v_now,
        revision=revision+1 WHERE installation_operation_id=
        v_setup.installation_operation_id;
    UPDATE iam_sessions SET state='revoked',revoked_at=v_now,
        revision=revision+1,updated_at=v_now
    WHERE user_id=v_setup.user_id AND state='active';
    UPDATE iam_refresh_families SET state='revoked',
        revoke_reason='initial_admin_setup',revoked_at=v_now,
        revision=revision+1,updated_at=v_now
    WHERE user_id=v_setup.user_id AND state='active';
    UPDATE iam_refresh_tokens SET state='revoked',revoked_at=v_now,
        revision=revision+1,updated_at=v_now
    WHERE family_id IN (SELECT id FROM iam_refresh_families
        WHERE user_id=v_setup.user_id) AND state='active';
    SELECT result.token_version,result.authorization_version
    INTO STRICT v_token_version,v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id,'iam.initial_admin_setup.accept',p_request_sha256,
        v_setup.user_id,v_setup.user_id,v_credential_revision,
        'credential_target',v_setup.user_id,'password_credential',true,true
    ) AS result;
    PERFORM platform.tenancy_customer_setup_accepted_v1(
        v_setup.installation_operation_id,p_tenant_id,v_setup.artifact_id
    );
    RETURN QUERY SELECT v_setup.user_id,v_setup.user_id,
        v_credential_revision,v_token_version,v_authorization_version,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_admin_setup_accept_v1(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_accept_v1(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea,bytea,bytea
) TO harpy_runtime;
