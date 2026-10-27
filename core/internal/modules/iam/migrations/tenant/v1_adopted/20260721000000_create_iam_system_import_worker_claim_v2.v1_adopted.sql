-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_claim_v2(
    p_scope_schema name,p_tenant_id uuid,p_import_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint
)
RETURNS TABLE(
    actor_id uuid,resource_kind text,authorization_version bigint,
    duplicate_action text,mapping_document bytea,next_row_index bigint,
    processed_rows bigint,source_id uuid,source_format text,object_key text,
    object_version text,header_sha256 bytea,ciphertext_sha256 bytea,
    ciphertext_bytes bigint,plaintext_sha256 bytea,plaintext_bytes bigint,
    key_id text,wrapped_dek bytea,wrap_nonce bytea,provider_request_id text,
    uploaded_at timestamptz,finished boolean,catalog_version_id uuid,
    catalog_sha256 bytea,identity_manifest_sha256 bytea
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_actor_id uuid;
    v_resource_kind text;
    v_authorization_version bigint;
    v_duplicate_action text;
    v_mapping_document bytea;
    v_next_row_index bigint;
    v_processed_rows bigint;
    v_source_id uuid;
    v_source_format text;
    v_object_key text;
    v_object_version text;
    v_header_sha256 bytea;
    v_ciphertext_sha256 bytea;
    v_ciphertext_bytes bigint;
    v_plaintext_sha256 bytea;
    v_plaintext_bytes bigint;
    v_key_id text;
    v_wrapped_dek bytea;
    v_wrap_nonce bytea;
    v_provider_request_id text;
    v_uploaded_at timestamptz;
    v_finished boolean;
    v_catalog_version_id uuid;
    v_catalog_sha256 bytea;
    v_identity_manifest_sha256 bytea;
BEGIN
    SELECT claim_result.actor_id,claim_result.resource_kind,
           claim_result.authorization_version,claim_result.duplicate_action,
           claim_result.mapping_document,claim_result.next_row_index,
           claim_result.processed_rows,claim_result.source_id,
           claim_result.source_format,claim_result.object_key,
           claim_result.object_version,claim_result.header_sha256,
           claim_result.ciphertext_sha256,claim_result.ciphertext_bytes,
           claim_result.plaintext_sha256,claim_result.plaintext_bytes,
           claim_result.key_id,claim_result.wrapped_dek,
           claim_result.wrap_nonce,claim_result.provider_request_id,
           claim_result.uploaded_at,claim_result.finished
    INTO STRICT v_actor_id,v_resource_kind,v_authorization_version,
                v_duplicate_action,v_mapping_document,v_next_row_index,
                v_processed_rows,v_source_id,v_source_format,v_object_key,
                v_object_version,v_header_sha256,v_ciphertext_sha256,
                v_ciphertext_bytes,v_plaintext_sha256,v_plaintext_bytes,
                v_key_id,v_wrapped_dek,v_wrap_nonce,v_provider_request_id,
                v_uploaded_at,v_finished
    FROM iam_system_record_import_worker_claim_v1(
        p_scope_schema,p_tenant_id,p_import_id,p_job_id,p_attempt_id,
        p_lease_owner_id,p_lease_generation
    ) AS claim_result;

    SELECT request.catalog_version_id,request.catalog_sha256,
           request.identity_manifest_sha256
    INTO STRICT v_catalog_version_id,v_catalog_sha256,
                v_identity_manifest_sha256
    FROM iam_system_record_imports AS request
    WHERE request.import_id=p_import_id
      AND request.tenant_id=p_tenant_id
      AND request.job_id=p_job_id
      AND request.actor_id=v_actor_id
      AND request.resource_kind=v_resource_kind;

    RETURN QUERY SELECT
        v_actor_id,v_resource_kind,v_authorization_version,v_duplicate_action,
        v_mapping_document,v_next_row_index,v_processed_rows,v_source_id,
        v_source_format,v_object_key,v_object_version,v_header_sha256,
        v_ciphertext_sha256,v_ciphertext_bytes,v_plaintext_sha256,
        v_plaintext_bytes,v_key_id,v_wrapped_dek,v_wrap_nonce,
        v_provider_request_id,v_uploaded_at,v_finished,v_catalog_version_id,
        v_catalog_sha256,v_identity_manifest_sha256;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_claim_v2(
    name,uuid,uuid,uuid,uuid,uuid,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_claim_v2(
    name,uuid,uuid,uuid,uuid,uuid,bigint
) TO harpy_runtime;
