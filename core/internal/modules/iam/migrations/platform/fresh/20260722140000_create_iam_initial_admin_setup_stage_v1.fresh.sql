-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_initial_admin_setup_stage_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='initial admin setup stage mutation is forbidden';
END;
$function$;

CREATE TABLE platform.iam_initial_admin_setup_stages (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    identifier_id uuid NOT NULL,
    normalized_email text NOT NULL,
    token_id uuid NOT NULL UNIQUE,
    selector bytea NOT NULL UNIQUE,
    hash_key_id text NOT NULL,
    secret_hmac bytea NOT NULL,
    binding_sha256 bytea NOT NULL,
    artifact_id uuid NOT NULL UNIQUE,
    artifact_ref text NOT NULL UNIQUE,
    kms_key_id text NOT NULL,
    nonce bytea NOT NULL,
    ciphertext bytea NOT NULL,
    authentication_tag bytea NOT NULL,
    aad_sha256 bytea NOT NULL,
    ciphertext_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_initial_admin_setup_stage_shape_check CHECK (
        normalized_email = pg_catalog.lower(normalized_email)
        AND pg_catalog.octet_length(normalized_email) BETWEEN 3 AND 254
        AND pg_catalog.octet_length(selector) = 16
        AND hash_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$'
        AND pg_catalog.octet_length(secret_hmac) = 32
        AND pg_catalog.octet_length(binding_sha256) = 32
        AND artifact_ref =
            'iam-lifecycle-delivery://tenant/' || tenant_id::text ||
            '/artifact/' || artifact_id::text
        AND pg_catalog.octet_length(nonce) = 12
        AND pg_catalog.octet_length(ciphertext) = 66
        AND pg_catalog.octet_length(authentication_tag) = 16
        AND pg_catalog.octet_length(aad_sha256) = 32
        AND pg_catalog.octet_length(ciphertext_sha256) = 32
        AND pg_catalog.isfinite(created_at)
    )
);

CREATE TRIGGER iam_initial_admin_setup_stages_immutable
BEFORE UPDATE OR DELETE ON platform.iam_initial_admin_setup_stages
FOR EACH ROW EXECUTE FUNCTION platform.iam_initial_admin_setup_stage_guard_v1();

CREATE FUNCTION platform.iam_initial_admin_setup_stage_v1(
    p_operation_id uuid, p_tenant_id uuid, p_user_id uuid,
    p_identifier_id uuid, p_normalized_email text, p_token_id uuid,
    p_selector bytea, p_hash_key_id text, p_secret_hmac bytea,
    p_binding_sha256 bytea, p_artifact_id uuid, p_artifact_ref text,
    p_kms_key_id text, p_nonce bytea, p_ciphertext bytea,
    p_authentication_tag bytea, p_aad_sha256 bytea,
    p_ciphertext_sha256 bytea, p_created_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.iam_initial_admin_setup_stages%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-api'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_user_id IS NULL OR p_identifier_id IS NULL
       OR p_normalized_email IS DISTINCT FROM pg_catalog.lower(p_normalized_email)
       OR pg_catalog.octet_length(p_normalized_email) NOT BETWEEN 3 AND 254
       OR p_token_id IS NULL OR pg_catalog.octet_length(p_selector) <> 16
       OR p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$'
       OR pg_catalog.octet_length(p_secret_hmac) <> 32
       OR pg_catalog.octet_length(p_binding_sha256) <> 32
       OR p_artifact_id IS NULL
       OR p_artifact_ref IS DISTINCT FROM
          'iam-lifecycle-delivery://tenant/' || p_tenant_id::text ||
          '/artifact/' || p_artifact_id::text
       OR pg_catalog.octet_length(p_nonce) <> 12
       OR pg_catalog.octet_length(p_ciphertext) <> 66
       OR pg_catalog.octet_length(p_authentication_tag) <> 16
       OR pg_catalog.octet_length(p_aad_sha256) <> 32
       OR pg_catalog.octet_length(p_ciphertext_sha256) <> 32
       OR NOT pg_catalog.isfinite(p_created_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='initial admin setup stage is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM platform.iam_initial_admin_setup_stages AS stage
    WHERE stage.operation_id=p_operation_id
    FOR SHARE;
    IF FOUND THEN
        IF ROW(v_existing.tenant_id,v_existing.user_id,
            v_existing.identifier_id,v_existing.normalized_email)
           IS DISTINCT FROM ROW(p_tenant_id,p_user_id,p_identifier_id,
            p_normalized_email) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='initial admin setup stage conflicts';
        END IF;
        RETURN true;
    END IF;
    INSERT INTO platform.iam_initial_admin_setup_stages (
        operation_id,tenant_id,user_id,identifier_id,normalized_email,
        token_id,selector,hash_key_id,secret_hmac,binding_sha256,
        artifact_id,artifact_ref,kms_key_id,nonce,ciphertext,
        authentication_tag,aad_sha256,ciphertext_sha256,created_at
    ) VALUES (
        p_operation_id,p_tenant_id,p_user_id,p_identifier_id,p_normalized_email,
        p_token_id,p_selector,p_hash_key_id,p_secret_hmac,p_binding_sha256,
        p_artifact_id,p_artifact_ref,p_kms_key_id,p_nonce,p_ciphertext,
        p_authentication_tag,p_aad_sha256,p_ciphertext_sha256,p_created_at
    );
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_v1(
    uuid,uuid,uuid,uuid,text,uuid,bytea,text,bytea,bytea,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_guard_v1()
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_initial_admin_setup_stage_v1(
    uuid,uuid,uuid,uuid,text,uuid,bytea,text,bytea,bytea,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea,timestamptz
) TO harpy_runtime;
