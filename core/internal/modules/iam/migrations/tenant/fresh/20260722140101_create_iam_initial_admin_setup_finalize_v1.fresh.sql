-- +goose Up
-- Migration stream version: 20260722140101.
-- +goose StatementBegin
CREATE FUNCTION iam_initial_admin_setup_finalize_v1(
    p_scope_schema name, p_operation_id uuid, p_tenant_id uuid,
    p_manifest_sha256 text
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage jsonb;
    v_stage_user_id uuid;
    v_stage_identifier_id uuid;
    v_stage_normalized_email text;
    v_stage_token_id uuid;
    v_stage_selector bytea;
    v_stage_hash_key_id text;
    v_stage_secret_hmac bytea;
    v_stage_binding_sha256 bytea;
    v_stage_artifact_id uuid;
    v_stage_artifact_ref text;
    v_stage_kms_key_id text;
    v_stage_nonce bytea;
    v_stage_ciphertext bytea;
    v_stage_authentication_tag bytea;
    v_stage_aad_sha256 bytea;
    v_stage_ciphertext_sha256 bytea;
    v_existing iam_initial_admin_setups%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_expires timestamptz;
BEGIN
    v_expires := v_now + INTERVAL '24 hours';
    v_stage := platform.iam_initial_admin_setup_stage_read_v1(
        p_operation_id,p_tenant_id
    );
    v_stage_user_id := (v_stage->>'user_id')::uuid;
    v_stage_identifier_id := (v_stage->>'identifier_id')::uuid;
    v_stage_normalized_email := v_stage->>'normalized_email';
    v_stage_token_id := (v_stage->>'token_id')::uuid;
    v_stage_selector := (v_stage->>'selector')::bytea;
    v_stage_hash_key_id := v_stage->>'hash_key_id';
    v_stage_secret_hmac := (v_stage->>'secret_hmac')::bytea;
    v_stage_binding_sha256 := (v_stage->>'binding_sha256')::bytea;
    v_stage_artifact_id := (v_stage->>'artifact_id')::uuid;
    v_stage_artifact_ref := v_stage->>'artifact_ref';
    v_stage_kms_key_id := v_stage->>'kms_key_id';
    v_stage_nonce := (v_stage->>'nonce')::bytea;
    v_stage_ciphertext := (v_stage->>'ciphertext')::bytea;
    v_stage_authentication_tag := (v_stage->>'authentication_tag')::bytea;
    v_stage_aad_sha256 := (v_stage->>'aad_sha256')::bytea;
    v_stage_ciphertext_sha256 := (v_stage->>'ciphertext_sha256')::bytea;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR NOT platform.tenancy_activation_assert_v2(
           p_operation_id,p_tenant_id,p_scope_schema
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='initial admin setup finalization denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM iam_installation_identity_facts AS fact
        JOIN iam_users AS app_user ON app_user.id=fact.first_user_id
        JOIN iam_user_identifiers AS identifier
          ON identifier.id=v_stage_identifier_id
         AND identifier.user_id=app_user.id
        WHERE fact.operation_id=p_operation_id
          AND fact.tenant_id=p_tenant_id
          AND fact.first_user_id=v_stage_user_id
          AND fact.normalized_email=v_stage_normalized_email
          AND app_user.status='active' AND app_user.identity_kind='human'
          AND identifier.kind='email_v1'
          AND identifier.normalized_value=v_stage_normalized_email
          AND identifier.state='active' AND identifier.revision=1
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='initial admin setup identity conflicts';
    END IF;
    SELECT * INTO v_existing
    FROM iam_initial_admin_setups AS setup
    WHERE setup.installation_operation_id=p_operation_id
    FOR SHARE;
    IF FOUND THEN
        IF v_existing.user_id IS DISTINCT FROM v_stage_user_id
           OR v_existing.token_id IS DISTINCT FROM v_stage_token_id
           OR v_existing.artifact_id IS DISTINCT FROM v_stage_artifact_id
           OR v_existing.generation IS DISTINCT FROM 1::bigint THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='initial admin setup replay conflicts';
        END IF;
        RETURN v_existing.artifact_id;
    END IF;
    INSERT INTO iam_one_time_tokens (
        id,purpose,user_id,identifier_id,identifier_kind,
        identifier_revision,portal_binding_id,portal_setup_revision,
        selector,hash_key_id,secret_hmac,binding_sha256,state,
        attempt_count,revision,created_at,updated_at,expires_at,purge_after
    ) VALUES (
        v_stage_token_id,'initial_admin_setup.v1',v_stage_user_id,
        v_stage_identifier_id,'email_v1',1,NULL,NULL,v_stage_selector,
        v_stage_hash_key_id,v_stage_secret_hmac,v_stage_binding_sha256,
        'pending',0,1,v_now,v_now,v_expires,v_expires + INTERVAL '24 hours'
    );
    INSERT INTO iam_lifecycle_delivery_artifacts (
        id,purpose,subject_user_id,resource_id,recipient_identifier_id,
        token_id,artifact_ref,kms_key_id,nonce,ciphertext,
        authentication_tag,aad_sha256,ciphertext_sha256,revision,
        created_at,expires_at
    ) VALUES (
        v_stage_artifact_id,'initial_admin_setup.v1',v_stage_user_id,
        v_stage_identifier_id,v_stage_identifier_id,v_stage_token_id,
        v_stage_artifact_ref,v_stage_kms_key_id,v_stage_nonce,
        v_stage_ciphertext,v_stage_authentication_tag,v_stage_aad_sha256,
        v_stage_ciphertext_sha256,1,v_now,v_expires
    );
    INSERT INTO iam_initial_admin_setups (
        installation_operation_id,user_id,token_id,artifact_id,generation,
        state,revision,created_at,accepted_at,revoked_at
    ) VALUES (
        p_operation_id,v_stage_user_id,v_stage_token_id,v_stage_artifact_id,
        1,'pending',1,v_now,NULL,NULL
    );
    PERFORM platform.tenancy_customer_setup_artifact_bind_v1(
        p_operation_id,p_tenant_id,v_stage_artifact_id
    );
    RETURN v_stage_artifact_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_admin_setup_finalize_v1(
    name,uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v1(
    name,uuid,uuid,text
) TO harpy_runtime;
