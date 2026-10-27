-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_effect_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_effect_id uuid,p_effect_receipt_id uuid,
    p_object_key text,p_key_id text,p_header_sha256 bytea,
    p_ciphertext_sha256 bytea,p_ciphertext_bytes bigint,p_wrapped_dek bytea,
    p_wrap_nonce bytea,p_request_sha256 bytea,p_started_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_replayed boolean;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF p_tenant_id IS DISTINCT FROM (
        SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
    ) OR pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_receipt_id=p_effect_id OR
       pg_catalog.octet_length(p_object_key) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 2148532224 OR
       pg_catalog.octet_length(p_wrapped_dek) NOT BETWEEN 32 AND 1024 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_started_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export effect authority is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.token_version=p_token_version
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND OR v_request.state<>'running' OR
       v_request.attempt_generation<>p_lease_generation THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export effect owner is stale';
    END IF;
    SELECT platform.platform_async_iam_export_effect_start_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,p_effect_id,p_effect_receipt_id,
        p_request_sha256,p_started_at
    ) INTO v_replayed;
    IF v_request.effect_started_at IS NOT NULL THEN
        IF v_replayed AND v_request.effect_id=p_effect_id AND
           v_request.effect_lease_owner_id=p_lease_owner_id AND
           v_request.effect_receipt_id=p_effect_receipt_id AND
           v_request.effect_object_key=p_object_key AND
           v_request.key_id=p_key_id AND
           v_request.header_sha256=p_header_sha256 AND
           v_request.ciphertext_sha256=p_ciphertext_sha256 AND
           v_request.ciphertext_bytes=p_ciphertext_bytes AND
           v_request.wrapped_dek=p_wrapped_dek AND
           v_request.wrap_nonce=p_wrap_nonce AND
           v_request.effect_request_sha256=p_request_sha256 AND
           v_request.effect_started_at=p_started_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export effect replay differs';
    END IF;
    IF v_replayed OR p_effect_id<>p_attempt_id OR
       p_started_at<v_request.started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export effect state conflicts';
    END IF;
    UPDATE iam_system_record_exports SET effect_id=p_effect_id,
        effect_lease_owner_id=p_lease_owner_id,
        effect_receipt_id=p_effect_receipt_id,effect_object_key=p_object_key,
        key_id=p_key_id,header_sha256=p_header_sha256,
        ciphertext_sha256=p_ciphertext_sha256,ciphertext_bytes=p_ciphertext_bytes,
        wrapped_dek=p_wrapped_dek,wrap_nonce=p_wrap_nonce,
        effect_request_sha256=p_request_sha256,effect_started_at=p_started_at
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_effect_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    uuid,uuid,text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz
) FROM PUBLIC;
