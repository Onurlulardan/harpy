-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION iam_email_verification_start_v2(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_token_id uuid, p_identifier_id uuid,
    p_selector bytea, p_hash_key_id text, p_secret_hmac bytea,
    p_binding_sha256 bytea, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_artifact_id uuid, p_subject_user_id uuid, p_resource_id uuid,
    p_recipient_identifier_id uuid, p_artifact_ref text, p_kms_key_id text,
    p_nonce bytea, p_ciphertext bytea, p_authentication_tag bytea,
    p_aad_sha256 bytea, p_ciphertext_sha256 bytea
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_result_id uuid; v_subject_user_id uuid; v_result_revision bigint;
    v_token_version bigint; v_authorization_version bigint;
    v_replayed boolean; v_insert bigint;
BEGIN
    SELECT started.result_id, started.subject_user_id, started.result_revision,
           started.token_version, started.authorization_version, started.replayed
    INTO v_result_id, v_subject_user_id, v_result_revision,
         v_token_version, v_authorization_version, v_replayed
    FROM iam_email_verification_start_v1(
        p_scope_schema, p_tenant_id, p_operation_id, p_request_sha256,
        p_token_id, p_identifier_id, p_selector, p_hash_key_id,
        p_secret_hmac, p_binding_sha256, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS started;
    IF NOT FOUND THEN RETURN; END IF;
    IF NOT v_replayed THEN
        v_insert := iam_lifecycle_delivery_artifact_insert_v1(
            p_scope_schema, p_tenant_id, p_artifact_id,
            'email_verification_v1', p_subject_user_id, p_resource_id,
            p_recipient_identifier_id, p_token_id, p_artifact_ref,
            p_kms_key_id, p_nonce, p_ciphertext, p_authentication_tag,
            p_aad_sha256, p_ciphertext_sha256
        );
    END IF;
    RETURN QUERY SELECT v_result_id, v_subject_user_id,
        v_result_revision, v_token_version,
        v_authorization_version, v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_email_verification_start_v2(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint,uuid,uuid,uuid,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea
) FROM PUBLIC;
