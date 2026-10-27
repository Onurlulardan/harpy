-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_delivery_artifact_insert_v1(
    p_scope_schema name, p_tenant_id uuid, p_artifact_id uuid,
    p_purpose text, p_subject_user_id uuid, p_resource_id uuid,
    p_recipient_identifier_id uuid, p_token_id uuid, p_artifact_ref text,
    p_kms_key_id text, p_nonce bytea, p_ciphertext bytea,
    p_authentication_tag bytea, p_aad_sha256 bytea,
    p_ciphertext_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_identifier_id uuid;
    v_portal_binding_id uuid;
    v_state text;
    v_expires_at timestamptz;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_artifact_id IS NULL OR p_subject_user_id IS NULL OR
       p_resource_id IS NULL OR p_recipient_identifier_id IS NULL OR
       p_token_id IS NULL OR
       p_purpose NOT IN (
           'email_verification_v1', 'password_reset_v1', 'portal_setup_v1'
       ) OR p_artifact_ref IS DISTINCT FROM
           'iam-lifecycle-delivery://tenant/' || p_tenant_id::text ||
           '/artifact/' || p_artifact_id::text OR
       pg_catalog.char_length(p_kms_key_id) > 512 OR p_kms_key_id !~
           '^gcp-kms://projects/([a-z][a-z0-9-]{4,28}[a-z0-9]|[0-9]{6,30})/locations/[a-z0-9][a-z0-9-]{0,62}/keyRings/[A-Za-z0-9_-]{1,63}/cryptoKeys/[A-Za-z0-9_-]{1,63}/cryptoKeyVersions/[1-9][0-9]{0,18}$' OR
       pg_catalog.octet_length(p_nonce) <> 12 OR
       pg_catalog.octet_length(p_ciphertext) <> 66 OR
       pg_catalog.octet_length(p_authentication_tag) <> 16 OR
       pg_catalog.octet_length(p_aad_sha256) <> 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) <> 32 OR
       NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id
             AND tenant.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM lifecycle delivery artifact input is invalid';
    END IF;
    SELECT token.user_id, token.identifier_id, token.portal_binding_id,
           token.state, token.expires_at
    INTO v_user_id, v_identifier_id, v_portal_binding_id,
         v_state, v_expires_at
    FROM iam_one_time_tokens AS token
    WHERE token.id = p_token_id AND token.purpose = p_purpose
    FOR UPDATE;
    IF NOT FOUND OR v_state <> 'pending' OR v_expires_at <= v_now OR
       v_user_id IS DISTINCT FROM p_subject_user_id THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM lifecycle delivery token changed';
    END IF;
    IF p_purpose = 'email_verification_v1' THEN
        IF p_resource_id IS DISTINCT FROM v_identifier_id OR
           p_recipient_identifier_id IS DISTINCT FROM v_identifier_id OR
           NOT EXISTS (
               SELECT 1 FROM iam_user_identifiers AS recipient
               WHERE recipient.id = p_recipient_identifier_id
                 AND recipient.user_id = v_user_id
                 AND recipient.kind = 'email_v1'
                 AND recipient.state IN ('pending', 'verified')
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001';
        END IF;
    ELSIF p_purpose = 'password_reset_v1' THEN
        IF p_resource_id IS DISTINCT FROM v_identifier_id OR
           NOT EXISTS (
               SELECT 1 FROM iam_user_identifiers AS recipient
               WHERE recipient.id = p_recipient_identifier_id
                 AND recipient.user_id = v_user_id
                 AND recipient.kind = 'email_v1'
                 AND recipient.state = 'active'
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001';
        END IF;
    ELSIF p_resource_id IS DISTINCT FROM v_portal_binding_id OR
       NOT EXISTS (
           SELECT 1 FROM iam_user_identifiers AS recipient
           WHERE recipient.id = p_recipient_identifier_id
             AND recipient.user_id = v_user_id
             AND recipient.kind = 'email_v1'
             AND recipient.state = 'active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001';
    END IF;
    INSERT INTO iam_lifecycle_delivery_artifacts (
        id, purpose, subject_user_id, resource_id,
        recipient_identifier_id, token_id, artifact_ref, kms_key_id,
        nonce, ciphertext, authentication_tag, aad_sha256,
        ciphertext_sha256, revision, created_at, expires_at
    ) VALUES (
        p_artifact_id, p_purpose, p_subject_user_id, p_resource_id,
        p_recipient_identifier_id, p_token_id, p_artifact_ref, p_kms_key_id,
        p_nonce, p_ciphertext, p_authentication_tag, p_aad_sha256,
        p_ciphertext_sha256, 1, v_now, v_expires_at
    );
    RETURN 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_delivery_artifact_insert_v1(
    name,uuid,uuid,text,uuid,uuid,uuid,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea
) FROM PUBLIC;
