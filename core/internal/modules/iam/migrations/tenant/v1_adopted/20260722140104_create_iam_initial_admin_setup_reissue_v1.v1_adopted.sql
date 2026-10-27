-- +goose Up
-- Migration stream version: 20260722140104.
ALTER TABLE iam_initial_admin_setups
ADD COLUMN last_reissue_operation_id uuid;

-- +goose StatementBegin
CREATE FUNCTION iam_initial_admin_setup_reissue_v1(
    p_scope_schema name, p_operation_id uuid, p_reissue_operation_id uuid,
    p_tenant_id uuid
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
    v_setup iam_initial_admin_setups%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_expires timestamptz;
BEGIN
    v_expires := v_now + INTERVAL '24 hours';
    v_stage := platform.iam_initial_admin_setup_stage_read_v1(
        p_reissue_operation_id,p_tenant_id
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
            IS DISTINCT FROM 'harpy-v2:core-api'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_operation_id IS NULL OR p_reissue_operation_id IS NULL
       OR p_tenant_id IS NULL
       OR NOT platform.tenancy_activation_assert_v2(
           p_operation_id,p_tenant_id,p_scope_schema
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='initial admin setup reissue denied';
    END IF;
    SELECT * INTO v_setup
    FROM iam_initial_admin_setups AS setup
    WHERE setup.installation_operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_setup.state<>'pending'
       OR v_setup.user_id IS DISTINCT FROM v_stage_user_id
       OR NOT EXISTS (
           SELECT 1 FROM iam_users AS app_user
           JOIN iam_user_identifiers AS identifier
             ON identifier.id=v_stage_identifier_id
            AND identifier.user_id=app_user.id
           WHERE app_user.id=v_stage_user_id AND app_user.status='active'
             AND app_user.identity_kind='human' AND identifier.kind='email_v1'
             AND identifier.normalized_value=v_stage_normalized_email
             AND identifier.state='active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='initial admin setup reissue identity conflicts';
    END IF;
    IF v_setup.last_reissue_operation_id=p_reissue_operation_id
       AND v_setup.artifact_id=v_stage_artifact_id
       AND v_setup.token_id=v_stage_token_id THEN
        PERFORM platform.tenancy_customer_setup_reissue_v1(
            p_operation_id,p_reissue_operation_id,p_tenant_id,
            v_setup.artifact_id,v_stage_artifact_id
        );
        PERFORM platform.platform_async_initial_admin_setup_reissue_enqueue_v1(
            p_operation_id,p_reissue_operation_id,p_tenant_id,
            v_stage_artifact_id,v_stage_artifact_ref,p_scope_schema
        );
        RETURN v_stage_artifact_id;
    END IF;
    PERFORM platform.tenancy_customer_setup_reissue_v1(
        p_operation_id,p_reissue_operation_id,p_tenant_id,
        v_setup.artifact_id,v_stage_artifact_id
    );
    UPDATE iam_one_time_tokens
    SET state='revoked',revoked_at=v_now,revision=revision+1,updated_at=v_now
    WHERE id=v_setup.token_id AND state='pending';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='initial admin setup token reissue conflicts';
    END IF;
    DELETE FROM iam_lifecycle_delivery_artifacts
    WHERE id=v_setup.artifact_id;
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
    UPDATE iam_initial_admin_setups
    SET token_id=v_stage_token_id,artifact_id=v_stage_artifact_id,
        generation=generation+1,revision=revision+1,
        last_reissue_operation_id=p_reissue_operation_id
    WHERE installation_operation_id=p_operation_id;
    PERFORM platform.platform_async_initial_admin_setup_reissue_enqueue_v1(
        p_operation_id,p_reissue_operation_id,p_tenant_id,
        v_stage_artifact_id,v_stage_artifact_ref,p_scope_schema
    );
    RETURN v_stage_artifact_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_admin_setup_reissue_v1(
    name,uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_reissue_v1(
    name,uuid,uuid,uuid
) TO harpy_runtime;
